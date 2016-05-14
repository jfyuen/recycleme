# Recycle me [![Build Status](https://travis-ci.org/jfyuen/recycleme.svg?branch=master)](https://travis-ci.org/jfyuen/recycleme) [![Coverage Status](https://coveralls.io/repos/github/jfyuen/recycleme/badge.svg?branch=master)](https://coveralls.io/github/jfyuen/recycleme?branch=master)

A tool to check product based on bar code (EAN8 and EAN13 currently supported) and give information on how to recycle product waste and packaging.
The website using this tool is http://www.howtorecycle.me

Currently only France (Paris) is supported as a country/location to recycle product as I do not have enough experience with other countries or regions.

A very good french website to check where to throw away stuff: http://tri-recyclage.ecoemballages.fr/

More information for Paris: http://www.paris.fr/enquetedecheteries

## Product data

Product data are currently not stored locally but on some remote websites.
They are scrapped at each request because I only want to keep packaging information on the website and no the whole product description.
This however may change in the future.

The following websites are currently scrapped:
- http://openfoodfacts.org
- http://www.upcitemdb.com
- http://www.isbnsearch.org
- http://90.80.54.225
- https://starrymart.co.uk
- http://www.misterpharmaweb.com
- http://www.meddispar.fr/
- http://www.digit-eyes.com

Moreover, http://www.amazon.fr (on french portal) is supported via their Product Advertising API, given the following credentials are set in the environment:
- RECYCLEME_ACCESS_KEY
- RECYCLEME_SECRET_KEY
- RECYCLEME_ASSOCIATE_TAG

The Amazon fetcher is deactivated is any variable is missing.

Contributions are welcomed to support more websites or databases.

## Website

The website is supported and hosted on http://www.howtorecycle.me.
It features a minimal database with product and scrapping/link to some websites.
The database to connect to is set up with the `RECYCLEME_MONGO_URI` environment variable.

The following environment variables are necessary to receive mails from the app:
- RECYCLEME_MAIL_HOST
- RECYCLEME_MAIL_RECIPIENT
- RECYCLEME_MAIL_USERNAME
- RECYCLEME_MAIL_PASSWORD

On mobile phones, photo can be used to scan bar codes.

## Heroku deployment

The app itself is "heroku" ready. Just deploy it directly after setting up the database environment variable:
```bash
$ heroku login
$ heroku create
$ heroku push heroku master
$ heroku open
```
Mongolab service may be used for the database.

 
## Run in server mode

```bash
$ recycleme -server
```
It will connect to the database specified in the `RECYCLEME_MONGO_URI` environment variable.
The server listens on port 8080 by default. A port can be specified with the `-p` flag.

## Command line tool

The command line tool also need a database set up.

```bash
$ recycleme $EAN
```

For example:

```bash
$ recycleme 7613034383808
map[{Bac à couvercle jaune 1}:[{0 Boîte carton}] {Bac à couvercle vert 0}:[{1 Film plastique} {4 Nourriture}]]
```

For json:
```bash
$ recycleme -json 7613034383808
{"Bac à couvercle jaune":[{"Name":"Boîte carton"}],"Bac à couvercle vert":[{"Name":"Film plastique"},{"Name":"Nourriture"}]}
```

## Tests
Tests also need a mongodb database, it is specified by the `RECYCLEME_MONGO_TEST_URI` environment variable.

## Roadmap/TODO

- Add link to legal regulations in the country
- Add geoloc information/data as to where to find a type of product. i.e batteries, lamps, ... if the bin is not available in the building or locally, depending on country or location.
- Support more countries/regions