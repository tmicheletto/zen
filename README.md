# Zendesk Search

## Prerequisites
[Go 1.15](https://golang.org/doc/go1.15)

## Compiling

Run the following command in the root directory

```
go build
```

## Testing
To run the unit tests run the following command in the root directory

```
go test ./...
```

## Usage
### Search
To execute a search against the json files supplied run the following command in the root directory after compiling the code.

```
./zen search
```
You'll be prompted for the type of search and the field you wish to search on. 

The search results will display up to 10 results if your query is broad.

Partial matches will return results so if you enter part of the user's name or email address for example you will likely get a result.


### List Fields
To list the fields available to search on, run the following command.

```
./zen list-fields
```
You'll be prompted to select the type of search in order to list the relevant fields.


## Known Limitations

- Boolean fields are not currently searchable. They need to be indexed as strings but ran out of time to do this.

- Tickets are not searchable by the `_id` field. Although searching by `external_id`works and the configuration of the two fields looks the same searching by `_id`does not return results and again I ran out of time to resolve it.