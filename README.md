# Recycle me

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

Contributions are welcomed to support more websites or databases.

## Website

The website is supported and hosted on http://www.howtorecycle.me.
It features a database with product and scrapping/link to some database.


## Command line tool

```bash
$ recycleme -d ${DATADIR} $EAN
```
Replace `${DATADIR}` with the path to json directory.

For example:

```bash
$ recycleme -d ${DATADIR} 7613034383808
map[{Bac à couvercle jaune 1}:[{0 Boîte carton}] {Bac à couvercle vert 0}:[{1 Film plastique} {4 Nourriture}]]
```

For json:
```bash
$ recycleme -d ${DATADIR} -json 7613034383808
{"Bac à couvercle jaune":[{"Name":"Boîte carton"}],"Bac à couvercle vert":[{"Name":"Film plastique"},{"Name":"Nourriture"}]}
```


## Roadmap/TODO

- Allow saving data
- Migrate to a database
- Add/update materials and products API and web site
- Add link to legal regulations in the country
- Add geoloc information/data as to where to find a type of product. i.e batteries, lamps, ... if the bin is not available in the building or locally, depending on country or location.
- Support more countries/regions