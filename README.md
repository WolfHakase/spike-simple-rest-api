# spike-simple-rest-api
This is a test project to see how to create a simple REST API using gorilla/mux.

## What is a spike?
According to Wikipedia: 

`A spike is a product development method originating from extreme programming that uses the simplest possible program to explore potential solutions. It is used to determine how much work will be required to solve or work around a software issue. Typically, a "spike test" involves gathering additional information or testing for easily reproduced edge cases. The term is used in agile software development approaches like Scrum or Extreme Programming.`

In these spikes no use is made of TDD. In fact, the code is largely untested. 

The code isn't necessarily clean. 

It is merely a tool to understand the subject, so that the actual production code can more easily be tested and made clean.

An added benefit to spikes is that they don't contain anything worth stealing, therefore they're perfect to build a portfolio with.

## Gorilla MUX
For this project https://github.com/gorilla/mux has been used.

They are currently looking for new maintainers, so there is a (minor) risk in adopting it.

In a real life situation you'd probably not want to rely as heavily on the gorilla/mux implementation as is done in this spike. Normally you would hide the implementation details behind your own interface/wrapper combination. That'd make it trivial to swap the web framework you're using.

## How to run
To run this project you don't have to do anything special.

From Goland go to the `main.go` and use the hotkey `ctrl + shift + f10`.

Or if you don't use a fancy IDE just run it like you would run any other Go script.

## What is implemented?

This simple API exposes the following endpoints:
- `GET /ping` returns 'pong' on success
- `POST /items/{id}/duplicate` duplicates the item pointed at by {id}
- `GET /items/{id}` returns the item pointed at by {id}
- `DELETE /items/{id}` deletes the item pointed at by {id}
- `PUT /items/{id}` updated the item pointed at by {id}. Expects a body containing the new name and description.
- `POST /items/` create the item in the request body, with an auto-incremented ID
- `GET /items/` returns a list with all the items
- `/` returns a 404 error

Every request made is automatically logged through a middleware.

Cors is enabled.

Graceful shutdown is implemented on `ctrl+c` input.

## Postman

In the folder `/postman` you can find a json export for a collection to be used in Postman.
Using this collection all the requests that are implemented can be (manually) tested.