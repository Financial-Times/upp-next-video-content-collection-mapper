package main

import (
	"errors"
	"fmt"
	health "github.com/Financial-Times/go-fthealth/v1_1"
	"github.com/Financial-Times/message-queue-go-producer/producer"
	"github.com/Financial-Times/message-queue-gonsumer/consumer"
	"github.com/Financial-Times/service-status-go/gtg"
	"net/http"
)

const healthPath = "/__health"

type healthService struct {
	config *healthConfig
	checks []health.Check
}

type healthConfig struct {
	appSystemCode string
	appName       string
	port          string
	httpCl        *http.Client
	consumerConf  consumer.QueueConfig
	producerConf  producer.MessageProducerConfig
	panicGuide    string
}

func newHealthService(config *healthConfig) *healthService {
	service := &healthService{config: config}
	service.checks = []health.Check{
		service.queueCheck(),
	}
	return service
}

func (service *healthService) queueCheck() health.Check {
	return health.Check{
		ID:               "message-queue-proxy-reachable",
		BusinessImpact:   "Related content from published Next videos will not be processed, clients will not see them within content.",
		Name:             "Message Queue Proxy Reachable",
		PanicGuide:       service.config.panicGuide,
		Severity:         1,
		TechnicalSummary: "Message queue proxy is not reachable/healthy",
		Checker:          service.checkAggregateMessageQueueProxiesReachable,
	}
}

func (service *healthService) checkAggregateMessageQueueProxiesReachable() (string, error) {
	var errMsg string

	err := service.checkMessageQueueProxyReachable(service.config.producerConf.Addr, service.config.producerConf.Topic, service.config.producerConf.Authorization, service.config.producerConf.Queue)
	if err != nil {
		return err.Error(), fmt.Errorf("Health check for queue address %s, topic %s failed. Error: %s", service.config.producerConf.Addr, service.config.producerConf.Topic, err.Error())
	}

	for i := 0; i < len(service.config.consumerConf.Addrs); i++ {
		err := service.checkMessageQueueProxyReachable(service.config.consumerConf.Addrs[i], service.config.consumerConf.Topic, service.config.consumerConf.AuthorizationKey, service.config.consumerConf.Queue)
		if err == nil {
			return "Ok", nil
		}
		errMsg = errMsg + fmt.Sprintf("Health check for queue address %s, topic %s failed. Error: %s", service.config.consumerConf.Addrs[i], service.config.consumerConf.Topic, err.Error())
	}
	return errMsg, errors.New(errMsg)
}

func (service *healthService) checkMessageQueueProxyReachable(address string, topic string, authKey string, queue string) error {
	req, err := http.NewRequest("GET", address+"/topics", nil)
	if err != nil {
		logger.messageEvent(topic, fmt.Sprintf("Could not connect to proxy: %v", err.Error()))
		return err
	}
	if len(authKey) > 0 {
		req.Header.Add("Authorization", authKey)
	}
	if len(queue) > 0 {
		req.Host = queue
	}
	resp, err := service.config.httpCl.Do(req)
	if err != nil {
		logger.messageEvent(topic, fmt.Sprintf("Could not connect to proxy: %v", err.Error()))
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("Proxy returned status: %d", resp.StatusCode)
		return errors.New(errMsg)
	}
	return nil
}

func (service *healthService) gtgCheck() gtg.Status {
	for _, check := range service.checks {
		if _, err := check.Checker(); err != nil {
			return gtg.Status{GoodToGo: false, Message: err.Error()}
		}
	}
	return gtg.Status{GoodToGo: true}
}
