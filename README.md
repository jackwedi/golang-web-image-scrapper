# golang-web-image-scrapper

This app was supposed to be a crawler of website images but instead is a scrapper of website images.
It will download into a local directory the images contained in the given urls.


## Usage

### Running the server
#### Local
```bash
go run main.go
```

#### Docker
```bash
docker run -p your_port:1337 -it app
```
or this to run in the background (use **docker ps docker kill** to kill afterward)
```bash
docker run -p your_port:1337 -d app
```

### Interacting with the server
#### Scrapping images
Send get ***http://localhost:1337/*** queries, with following the body schema: 

```JSON
{
	"jobname" : "JOB1",
	"urls": 
		[
		    "https://coderwall.com/p/cp5fya/measuring-execution-time-in-go",
		    "https://na.leagueoflegends.com/fr-fr/",
		    "https://unsplash.com/"
		]
	
}
```

#### Metrics
Metrics are available in the ***http://localhost:1337/metrics*** query. The metrics are stored at runtime and not persisted. They gives information about the time spent for each job and the number of file dowloaded.

### TODOS

-  add total weight download to compare with theoric download speed
- save metrics somewhere
- use Prometheus?

#### Crawling
- Not actually crawling through the website, need to look through all urls in the website and download their images aswell
- Instead of saving locally the images, save them into a Amazon S3 or save URLS into a DB
- Output the not succedded image download
- Format the scrapped images urls name into file safe names instead of giving them an arbitrary name or discrding them
- App is crashing on errors, need to be handling errors by just logging them
