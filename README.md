# Recycle me

A tool to check product based on bar code (EAN8 and EAN13 currently supported) and give information on how to recycle product waste and packaging.
The website using this tool is http://www.howtorecycle.me

Currently only France (Paris) is supported as a country/location to recycle product as I do not have enough experience with other countries.

Very good french website to check where to throw away stuff: http://tri-recyclage.ecoemballages.fr/

## Product data

Product data are currently not stored locally buy on some remote websites.
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
$ recycleme $EAN
```

For example:

```bash
$ recycleme 7613034383808
Four à Pierre Royale (7613034383808) at http://fr.openfoodfacts.org/api/v0/produit/7613034383808.json
	Image: http://static.openfoodfacts.org/images/products/761/303/438/3808/front.8.400.jpg
```

## Roadmap/TODO

- Json usage
- Create the website
- Migrate to mongodb (or another db, sql?) when the volume will be sufficient
- Add geoloc information/data as to where to find a type of product. i.e batteries, lamps, ... if the bin is not available in the building or locally, depending on country or location.
- Add link to legal regulations in the country
- Support more countries

