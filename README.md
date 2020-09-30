<div align="center">
  <br/>
  <img src="./docs/images/logo-words.png" width="250"/>
  <br/>
  <br/>
  <p>
    Simple, distributed, lighweight<br>
    RESTful search engine implementation written in go
  </p>
  <p>
    <a href="https://github.com/actumn/searchgoose/blob/master/LICENSE">
      <img src="https://img.shields.io/badge/license-MIT-blue.svg"/>
    </a>
  </p>
</div>

---
## Motivation

Study purposes, mostly for understanding the implementation details of how
[elasticearch](https://github.com/elastic/elasticsearch) is built, focusing on clustering distributed system and supporting full-text search using [bleve](https://github.com/blevesearch/bleve).

## Run
```shell script
$ go run main.go
```

## API
### Create Index
```
PUT /test15
content-type: application/json

{
  "settings": {
    "number_of_shards": 3
  },
  "mappings": {
    "properties": {
      "field1": {
        "type": "text"
      }
    }
  }
}
```

### Document Index
```
PUT /test15/_doc/4
content-type: application/json

{
  "field1": "test",
  "field2": "test2"
}
```

### Search
```
POST /test15/_search
content-type: application/json

{
  "size": 100,
  "query": {
    "match": {
      "field1": "field test"
    } 
  }
}
```