# Gif Group API

This API is a basic solution for a game application where users create Gif Groups under a common theme. The API is written in Go and backed by Redis.
Responses are in JSON, and responds to the following endpoints:

- [GET] /groups - returns all groupings of gifs
- [GET] /groups/{id}/gifs - returns all gifs for the group matching the id specified
- [POST] /groups - creates a new group with name
- [POST] /groups/{id}/gifs - creates a new gif within the group matching the id specified

# Setup
[setup instructions once application is up and running]

# Response Format
Response format will be in JSON, and follow the structure below:
```json
{
    "success": true,
    "status_code": 200,
    "status_text": "OK",
    "error_code": 0,
    "error_text": "No error",
    "content": [ // ... array of response objects ]
}
```
# Endpoints

##### GET `/groups`
Returns all groupings of gifs.

##### GET `/groups/{id}/gifs`
Returns all gifs for grouping corresponding to the specified `{id}` parameter.

##### POST `/groups`
Creates a new gif grouping.

##### POST `/groups/{id}/gifs`
Creates a new gif within grouping corresponding to the specified `{id}` parameter.
