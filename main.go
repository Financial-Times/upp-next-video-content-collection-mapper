package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Financial-Times/go-logger/v2"
	"github.com/Financial-Times/kafka-client-go/v3"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	"github.com/gorilla/handlers"
	"github.com/jawher/mow.cli"
)

const serviceDescription = "Get the related content references from the Next video content, creates a story package holding those references and puts a message with them on kafka queue for further processing and ingestion on Neo4j."

type serviceConfig struct {
	appName     string
	serviceName string
	port        string
}

func main() {
	app := cli.App("next-video-content-collection-mapper", serviceDescription)

	appSystemCode := app.String(cli.StringOpt{
		Name:   "app-system-code",
		Value:  "upp-next-video-content-collection-mapper",
		Desc:   "System Code of the application",
		EnvVar: "APP_SYSTEM_CODE",
	})
	serviceName := app.String(cli.StringOpt{
		Name:   "service-name",
		Value:  "next-video-content-collection-mapper",
		Desc:   "The name of this service",
		EnvVar: "SERVICE_NAME",
	})
	appName := app.String(cli.StringOpt{
		Name:   "app-name",
		Value:  "next-video-content-collection-mapper",
		Desc:   "Application name",
		EnvVar: "APP_NAME",
	})
	port := app.String(cli.StringOpt{
		Name:   "port",
		Value:  "8080",
		Desc:   "Port to listen on",
		EnvVar: "APP_PORT",
	})
	panicGuide := app.String(cli.StringOpt{
		Name:   "panic-guide",
		Value:  "https://dewey.ft.com/upp-next-video-cc-mapper.html",
		Desc:   "Path to panic guide",
		EnvVar: "PANIC_GUIDE",
	})
	kafkaAddress := app.String(cli.StringOpt{
		Name:   "queue-kafkaAddress",
		Value:  "",
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "KAFKA_ADDRESS",
	})
	group := app.String(cli.StringOpt{
		Name:   "group",
		Value:  "NextVideoContentCollectionMapper",
		Desc:   "Group used to read the messages from the queue.",
		EnvVar: "Q_GROUP",
	})
	readTopic := app.String(cli.StringOpt{
		Name:   "read-topic",
		Value:  "NativeCmsPublicationEvents",
		Desc:   "The topic to read the messages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "CmsPublicationEvents",
		Desc:   "The topic to write the messages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	logLevel := app.String(cli.StringOpt{
		Name:   "logLevel",
		Value:  "INFO",
		Desc:   "Logging level {DEBUG, INFO, WARN, ERROR}",
		EnvVar: "LOG_LEVEL",
	})
	consumerLagTolerance := app.Int(cli.IntOpt{
		Name:   "consumerLagTolerance",
		Value:  120,
		Desc:   "Kafka lag tolerance",
		EnvVar: "KAFKA_LAG_TOLERANCE",
	})

	log := logger.NewUPPLogger(*serviceName, *logLevel)

	log.Infof("[Startup] %s is starting ", *serviceName)

	app.Action = func() {
		if len(*kafkaAddress) == 0 {
			log.Fatal("No queue address provided. Quitting...")
		}

		sc := serviceConfig{
			appName:     *appName,
			serviceName: *serviceName,
			port:        *port,
		}
		sh := serviceHandler{sc, log}

		consumerConfig := kafka.ConsumerConfig{
			BrokersConnectionString: *kafkaAddress,
			ConsumerGroup:           *group,
			ConnectionRetryInterval: time.Minute,
		}

		topics := []*kafka.Topic{
			kafka.NewTopic(*readTopic, kafka.WithLagTolerance(int64(*consumerLagTolerance))),
		}

		consumer := kafka.NewConsumer(consumerConfig, topics, log)

		producerConfig := kafka.ProducerConfig{
			BrokersConnectionString: *kafkaAddress,
			Topic:                   *writeTopic,
			ConnectionRetryInterval: time.Minute,
		}

		producer := kafka.NewProducer(producerConfig, log)
		defer func(producer *kafka.Producer) {
			err := producer.Close()
			if err != nil {
				log.WithError(err).Error("Producer could not stop")
			}
		}(producer)

		qh := queueHandler{
			sc:              sc,
			messageProducer: producer,
			log:             log}

		go consumer.Start(qh.queueConsume)
		defer func(consumer *kafka.Consumer) {
			err := consumer.Close()
			if err != nil {
				log.WithError(err).Error("Consumer could not stop")
			}
		}(consumer)

		hc := NewHealthCheck(producer, consumer, *appName, *appSystemCode, *panicGuide)

		go func() {
			serveAdminEndpoints(&sh, hc, log)
		}()

		waitForSignal()
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func serveAdminEndpoints(sh *serviceHandler, hc *HealthCheck, log *logger.UPPLogger) {
	serveMux := http.NewServeMux()

	serveMux.Handle("/map", handlers.MethodHandler{"POST": http.HandlerFunc(sh.mapRequest)})
	serveMux.HandleFunc("/__health", hc.Health())
	serveMux.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(hc.GTG))
	serveMux.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	log.Info("Service started", sh.sc.asMap())

	if err := http.ListenAndServe(":"+sh.sc.port, serveMux); err != nil {
		log.Fatalf("Unable to start: %v", err)
	}
}

func waitForSignal() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

func (sc serviceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"app-name":     sc.appName,
		"service-name": sc.serviceName,
		"service-port": sc.port,
	}
}
