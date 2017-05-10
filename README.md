# Next Video Content Collection Mapper
[![Circle CI](https://circleci.com/gh/Financial-Times/upp-next-video-content-collection-mapper/tree/master.png?style=shield)](https://circleci.com/gh/Financial-Times/upp-next-video-content-collection-mapper/tree/master)[![Go Report Card](https://goreportcard.com/badge/github.com/Financial-Times/upp-next-video-content-collection-mapper)](https://goreportcard.com/report/github.com/Financial-Times/upp-next-video-content-collection-mapper) [![Coverage Status](https://coveralls.io/repos/github/Financial-Times/upp-next-video-content-collection-mapper/badge.svg)](https://coveralls.io/github/Financial-Times/upp-next-video-content-collection-mapper)

## Introduction

Get the related content references from the Next video content ("related" field) by listening to kafka queue, creates a story package holding those references and puts a message with them on kafka queue for further processing and ingestion on Neo4j.

## Installation
      
Download the source code, dependencies and test dependencies:

        go get -u github.com/kardianos/govendor
        go get -u github.com/Financial-Times/upp-next-video-content-collection-mapper
        cd $GOPATH/src/github.com/Financial-Times/upp-next-video-content-collection-mapper
        govendor sync
        go build .

## Running locally

1. Run the tests and install the binary:

        govendor sync
        govendor test -v -race
        go install

2. Run the binary (using the `help` flag to see the available optional arguments):

        $GOPATH/bin/next-video-content-collection-mapper [--help]

Options:

        --app-system-code="upp-next-video-content-collection-mapper"    System Code of the application ($APP_SYSTEM_CODE)
        --app-name="Next Video Content Collection Mapper"               Application name ($APP_NAME)
        --service-name="next-video-content-collection-mapper"           Service name ($SERVICE_NAME)
        --port="8080"                                                   Port to listen on ($APP_PORT)
        --queue-addresses="http://%H:8080"                              Queue address ($Q_ADDR)
        --group="NextVideoContentCollectionMapper"                      Group used to read messages from queue ($Q_GROUP)
        --read-topic="NativeCmsPublicationEvents"                       Queue topic name from where to read the messages ($Q_READ_TOPIC)
        --read-queue="kafka"                                            The queue to read the messages from ($Q_READ_QUEUE)
        --write-topic="CmsPublicationEvents"                            Queue topic name where to write the messages ($Q_WRITE_TOPIC)
        --write-queue="kafka"                                           The queue to write the messages to ($Q_WRITE_QUEUE)

There are defaults values used for properties so when deployed locally it can be run the excutable only.

3. Test:

`
curl http://localhost:8080/__next-video-content-collection-mapper/__health | jq
`

## Build and deployment

* Built by Docker Hub on merge to master: [coco/next-video-content-collection-mapper](https://hub.docker.com/r/coco/upp-next-video-content-collection-mapper/)
* CI provided by CircleCI: [next-video-content-collection-mapper](https://circleci.com/gh/Financial-Times/upp-next-video-content-collection-mapper)

## Service endpoints

### POST

#### /map

Example:

`
curl -X POST http://localhost:8080/map -H "Content-Type: application/json" -H "X-Request-Id: tid_12345" -H "X-Origin-System-Id: next-video-editor" -d @body.json
`


body.json:
```
{
	"_id": "58d8d6cc789d4c000f6b0169",
	"updatedAt": "2017-04-03T16:30:11.106Z",
	"createdAt": "2017-03-27T09:09:32.541Z",
	"mioId": 762380,
	"title": "Trump trade under scrutiny",
	"createdBy": "seb.morton-clark",
	"encoding": {
		"job": 358759376,
		"status": "COMPLETE",
		"outputs": [{
			"audioCodec": "mp3",
			"duration": 65904,
			"mediaType": "audio/mpeg",
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/0x0.mp3"
		},
		{
			"audioCodec": "aac",
			"videoCodec": "h264",
			"duration": 65940,
			"mediaType": "video/mp4",
			"height": 360,
			"width": 640,
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/640x360.mp4"
		},
		{
			"audioCodec": "aac",
			"videoCodec": "h264",
			"duration": 65940,
			"mediaType": "video/mp4",
			"height": 720,
			"width": 1280,
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/1280x720.mp4"
		}]
	},
	"__v": 0,
	"byline": "Filmed by Niclola Stansfield. Produced by Seb Morton-Clark.",
	"description": "Global equities are on the defensive, led by weaker commodities and financials as investors scrutinise the viability of the Trump trade. The FT's Mike Mackenzie reports.",
	"image": "https://api.ft.com/content/ffc60243-2b77-439a-a6c9-0f3603ee5f83",
	"standfirst": "Mike Mackenzie provides analysis of the morning's market news",
	"updatedBy": "seb.morton-clark",
	"isPublished": false,
	"related": [{
		"uuid": "c4cde316-128c-11e7-80f4-13e067d5072c",
	}],
	"annotations": [{
		"id": "http://api.ft.com/things/059c58aa-53e6-306b-8642-e718c869ec09",
		"predicate": "http://www.ft.com/ontology/classification/isPrimarilyClassifiedBy"
	},
	{
		"id": "http://api.ft.com/things/d969d76e-f8f4-34ae-bc38-95cfd0884740",
		"predicate": "http://www.ft.com/ontology/classification/isClassifiedBy"
	},
	{
		"id": "http://api.ft.com/things/b43f1a91-b805-3453-8c36-1d164c047ca2",
		"predicate": "http://www.ft.com/ontology/annotation/mentions"
	},
	{
		"id": "http://api.ft.com/things/c2acc00f-5ad6-3f79-b6bb-d51dcd81508a",
		"predicate": "http://www.ft.com/ontology/annotation/about"
	},
	{
		"id": "http://api.ft.com/things/c2acc00f-5ad6-3f79-b6bb-d51dcd81508a",
		"predicate": "http://www.ft.com/ontology/annotation/majorMentions"
	}],
	"encodings": [{
		"mioId": 762380,
		"name": "Trump trade under scrutiny",
		"primary": true,
		"status": "COMPLETE",
		"job": 358759376,
		"outputs": [{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/0x0.mp3",
			"mediaType": "audio/mpeg",
			"duration": 65904,
			"audioCodec": "mp3"
		},
		{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/640x360.mp4",
			"width": 640,
			"height": 360,
			"mediaType": "video/mp4",
			"duration": 65940,
			"videoCodec": "h264",
			"audioCodec": "aac"
		},
		{
			"url": "http://ftvideo.prod.zencoder.outputs.s3.amazonaws.com/e2290d14-7e80-4db8-a715-949da4de9a07/1280x720.mp4",
			"width": 1280,
			"height": 720,
			"mediaType": "video/mp4",
			"duration": 65940,
			"videoCodec": "h264",
			"audioCodec": "aac"
		}]
	}],
	"canBeSyndicated": true,
	"transcription": {
		"status": "COMPLETE",
		"job": "1579674",
		"transcript": "<p>Here's what we're watching with trading underway in London. Global equities under pressure led by weaker commodities and financials as investors scrutinise the viability of the Trump trade. The dollar is weaker. Havens like yen, gold, and government bonds finding buyers. </p><p>As the dust settles over the failure to replace Obamacare, focus now on whether tax reform and other fiscal measures will eventuate. This is where the rubber meets the road for the Trump trade. High flying equity markets had been underpinned by the promise of big tax cuts and fiscal stimulus. And Wall Street is souring. </p><p>One big beneficiary of lower corporate taxes under Trump are small caps. They are now down 2 and 1/2% for the year. While the sector is still much higher since November, this is a key market barometer of prospects for the Trump trade. </p><p>Now while many still think some measure of tax reform or spending will eventuate, markets are very wary, namely of the risk that Congress and the Trump administration fail to reach agreement on legislation, that unlike health care reform, matters a great deal more to investors. </p><p>[MUSIC PLAYING] </p>",
		"captions": [{
			"format": "vtt",
			"url": "https://next-video-editor.ft.com/e2290d14-7e80-4db8-a715-949da4de9a07.vtt",
			"mediaType": "text/vtt"
		}]
	},
	"format": [],
	"type": "video",
	"id": "ad543253-ad5e-471c-a006-ebb395323028"
}
```

Response 200

Body:
```
{
	"payload": {
		"uuid": "151d4420-6ce6-3964-ad64-916561612973",
		"items": [{
			"uuid": "c4cde316-128c-11e7-80f4-13e067d5072c"
		}],
		"publishReference": "tid-12321123",
		"type": "story-package"
	},
	"contentUri": "http://next-video-content-collection-mapper.svc.ft.com/content-collection/story-package/151d4420-6ce6-3964-ad64-916561612973",
	"uuid": "151d4420-6ce6-3964-ad64-916561612973"
}
```

Response 400

If the mapping couldn't be performed because of invalid provided content.

## Healthchecks
Admin endpoints are:

`/__gtg`

`/__health`

`/__build-info`

Following check is performed for health and gtg endpoints:
* Checks that the connection to queue can be established.

### Logging

* The application uses [logrus](https://github.com/Sirupsen/logrus); the log file is initialised in [main.go](main.go).
* Logging requires an `env` app parameter, for all environments other than `local` logs are written to file.
* When running locally, logs are written to console. If you want to log locally to file, you need to pass in an env parameter that is != `local`.
* NOTE: `/__build-info` and `/__gtg` endpoints are not logged as they are called every second from varnish/vulcand and this information is not needed in logs/splunk.
