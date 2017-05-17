package main

import (
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	status "github.com/Financial-Times/service-status-go/httphandlers"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/handlers"
	"github.com/jawher/mow.cli"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const serviceDescription = "Get the related content references from the Next video content, creates a story package holding those references and puts a message with them on kafka queue for further processing and ingestion on Neo4j."

var logger *appLogger
var timeout = 10 * time.Second
var httpCl = &http.Client{Timeout: timeout}

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
	addresses := app.Strings(cli.StringsOpt{
		Name:   "queue-addresses",
		Value:  []string{"http://localhost:8080"},
		Desc:   "Addresses to connect to the queue (hostnames).",
		EnvVar: "Q_ADDR",
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
		Desc:   "The topic to read the meassages from.",
		EnvVar: "Q_READ_TOPIC",
	})
	readQueue := app.String(cli.StringOpt{
		Name:   "read-queue",
		Value:  "kafka",
		Desc:   "The queue to read the meassages from.",
		EnvVar: "Q_READ_QUEUE",
	})
	writeTopic := app.String(cli.StringOpt{
		Name:   "write-topic",
		Value:  "CmsPublicationEvents",
		Desc:   "The topic to write the meassages to.",
		EnvVar: "Q_WRITE_TOPIC",
	})
	writeQueue := app.String(cli.StringOpt{
		Name:   "write-queue",
		Value:  "kafka",
		Desc:   "The queue to write the meassages to.",
		EnvVar: "Q_WRITE_QUEUE",
	})

	logger = newAppLogger(*appName)

	app.Action = func() {
		if len(*addresses) == 0 {
			logger.log.Info("No queue address provided. Quitting...")
			cli.Exit(1)
		}

		sc := serviceConfig{
			appName:     *appName,
			serviceName: *serviceName,
			port:        *port,
		}
		sh := serviceHandler{sc}

		consumerConfig := consumer.QueueConfig{
			Addrs:                *addresses,
			Group:                *group,
			Topic:                *readTopic,
			Queue:                *readQueue,
			ConcurrentProcessing: false,
			AutoCommitEnable:     true,
		}

		producerConfig := producer.MessageProducerConfig{
			Addr:  (*addresses)[0],
			Topic: *writeTopic,
			Queue: *writeQueue,
		}

		hc := healthConfig{
			appSystemCode: *appSystemCode,
			appName:       *appName,
			port:          *port,
			httpCl:        httpCl,
			consumerConf:  consumerConfig,
			producerConf:  producerConfig,
		}

		go func() {
			serveAdminEndpoints(*appSystemCode, sc, sh, hc)
		}()

		qh := queueHandler{sc: sc, httpCl: httpCl, consumerConfig: consumerConfig, producerConfig: producerConfig}
		qh.init()

		consumeUntilSigterm(qh.messageConsumer, consumerConfig)
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("App could not start, error=[%s]\n", err)
		return
	}
}

func serveAdminEndpoints(appSystemCode string, sc serviceConfig, sh serviceHandler, hc healthConfig) {

	serveMux := http.NewServeMux()

	serveMux.Handle("/map", handlers.MethodHandler{"POST": http.HandlerFunc(sh.mapRequest)})

	healthService := newHealthService(&hc)
	h := health.HealthCheck{SystemCode: appSystemCode, Name: sc.appName, Description: serviceDescription, Checks: healthService.checks}
	serveMux.HandleFunc(healthPath, health.Handler(h))
	serveMux.HandleFunc(status.GTGPath, status.NewGoodToGoHandler(healthService.gtgCheck))
	serveMux.HandleFunc(status.BuildInfoPath, status.BuildInfoHandler)

	logger.serviceStartedEvent(sc.asMap())

	if err := http.ListenAndServe(":"+sc.port, serveMux); err != nil {
		log.Fatalf("Unable to start: %v", err)
	}
}

func consumeUntilSigterm(messageConsumer consumer.MessageConsumer, config consumer.QueueConfig) {
	logger.messageEvent(config.Topic, "Starting queue consumer")

	var consumerWaitGroup sync.WaitGroup
	consumerWaitGroup.Add(1)
	go func() {
		messageConsumer.Start()
		consumerWaitGroup.Done()
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	messageConsumer.Stop()
	consumerWaitGroup.Wait()
}

func (sc serviceConfig) asMap() map[string]interface{} {
	return map[string]interface{}{
		"app-name":     sc.appName,
		"service-name": sc.serviceName,
		"service-port": sc.port,
	}
}
