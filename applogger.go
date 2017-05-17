package main

import "github.com/Sirupsen/logrus"

type queueEvent struct {
	serviceName   string
	queueName     string
	queueTopic    string
	transactionID string
}

type appLogger struct {
	log         *logrus.Logger
	serviceName string
}

func newAppLogger(serviceName string) *appLogger {
	logrus.SetLevel(logrus.InfoLevel)
	log := logrus.New()
	log.Formatter = new(logrus.JSONFormatter)
	return &appLogger{log: log, serviceName: serviceName}
}

func (appLogger *appLogger) serviceStartedEvent(serviceConfig map[string]interface{}) {
	serviceConfig["event"] = "service_started"
	appLogger.log.WithFields(serviceConfig).Infof("%s started with configuration", appLogger.serviceName)
}

func (appLogger *appLogger) messageEvent(queueTopic string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":        "consume_queue",
		"service_name": appLogger.serviceName,
		"queue_topic":  queueTopic,
	}).Info(message)
}

func (appLogger *appLogger) warnMessageEvent(event queueEvent, videoUUID string, err error, errMessage string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   event.serviceName,
		"queue_name":     event.queueName,
		"queue_topic":    event.queueTopic,
		"transaction_id": event.transactionID,
		"uuid":           videoUUID,
		"error":          err,
	}).Warn(errMessage)
}

func (appLogger *appLogger) messageSentEvent(event queueEvent, uuid string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "produce_queue",
		"service_name":   event.serviceName,
		"queue_name":     event.queueName,
		"queue_topic":    event.queueTopic,
		"transaction_id": event.transactionID,
		"uuid":           uuid,
	}).Info(message)
}

func (appLogger *appLogger) videoEvent(transactionID string, videoUUID string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"uuid":           videoUUID,
	}).Warn(message)
}

func (appLogger *appLogger) videoMapEvent(transactionID string, videoUUID string, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "map",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"uuid":           videoUUID,
	}).Info(message)
}

func (appLogger *appLogger) videoErrorEvent(transactionID string, videoUUID string, err error, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"error":          err,
		"uuid":           videoUUID,
	}).Warn(message)
}

func (appLogger *appLogger) serviceEvent(transactionID string, err error, message string) {
	appLogger.log.WithFields(logrus.Fields{
		"event":          "error",
		"service_name":   appLogger.serviceName,
		"transaction_id": transactionID,
		"error":          err,
	}).Warn(message)
}
