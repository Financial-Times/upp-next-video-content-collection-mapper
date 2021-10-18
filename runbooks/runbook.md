# UPP - Next Video Content Collection Mapper

Next Video Content Collection Mapper references from the Next video content ("related" field) by listening to kafka queue,
creates a story package holding those references and puts a message with them on kafka queue for further processing and ingestion on Neo4j.

## Code

upp-next-video-cc-mapper

## Primary URL

<https://upp-prod-delivery-glb.upp.ft.com/__upp-next-video-cc-mapper/>

## Service Tier

Bronze

## Lifecycle Stage

Production

## Host Platform

AWS

## Architecture

Next Video Content Collection Mapper references from the Next video content ("related" field) by listening to kafka queue,
creates a story package holding those references and puts a message with them on kafka queue for further processing and ingestion on Neo4j.

## Contains Personal Data

No

## Contains Sensitive Data

No

## Failover Architecture Type

ActiveActive

## Failover Process Type

FullyAutomated

## Failback Process Type

FullyAutomated

## Failover Details

The service is deployed in both Delivery clusters. The failover guide for the cluster is located here:
<https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster>

## Data Recovery Process Type

NotApplicable

## Data Recovery Details

The service does not store data, so it does not require any data recovery steps.

## Release Process Type

PartiallyAutomated

## Rollback Process Type

Manual

## Release Details

Manual failover is needed when a new version of the service is deployed to production. Otherwise, an automated failover is going to take place when releasing. For more details about the failover process please see: <https://github.com/Financial-Times/upp-docs/tree/master/failover-guides/delivery-cluster>

## Key Management Process Type

Manual

## Key Management Details

To access the service clients need to provide basic auth credentials.
To rotate credentials you need to login to a particular cluster and update varnish-auth secrets.

## Monitoring

Service in UPP K8S delivery clusters:

- Delivery-Prod-EU health: <https://upp-prod-delivery-eu.upp.ft.com/__health/__pods-health?service-name=upp-next-video-content-collection-mapper>
- Delivery-Prod-US health: <https://upp-prod-delivery-us.upp.ft.com/__health/__pods-health?service-name=upp-next-video-content-collection-mapper>

## First Line Troubleshooting

<https://github.com/Financial-Times/upp-docs/tree/master/guides/ops/first-line-troubleshooting>

## Second Line Troubleshooting

Please refer to the GitHub repository README for troubleshooting information.
