`crawl` defined the pipeline's processing stage from source to destination that will scrap the html page for link, text , and inserted also indexed to persistence layer.
* 1st stage it will attempts to retrieve the contents of link by sending out HTTP GET requests
* 2nd stage extract every pointed out link that existed from preceding stage
* 3rd stage extract the content (text) of the page.
* 4th stage it will brodcast (duplicate) the payload to persisted, there will be 2 data persistence 1st for store the link(url) and 2nd for indexing or calculate the rank.