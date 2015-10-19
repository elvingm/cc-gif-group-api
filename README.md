# Gif Group API

This API is a basic solution for a game application where users create Gif Groups under a common theme. The API is written in Go and backed by Redis.
Responses are in JSON, and responds to the following endpoints:

*Routes are prefixed with `/api/v{version_number}*

- [GET] /api//groups - returns all groupings of gifs
- [GET] /groups/{id}/gifs - returns all gifs for the group matching the id specified
- [POST] /groups - creates a new group with name
- [POST] /groups/{id}/gifs - creates a new gif within the group matching the id specified

# Setup
In order to get the api running locally:

 1. `git clone` this repo
 2. `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` need to be provided the proper keys for development
 3. `cd cc-gifgroup-api`
 4. `godep go install`
 5. `cc-gifgroup-api`
 6. `curl http://localhost:1323/groups`

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
e.g. `curl http://localhost:1323/api/v1/groups`

##### GET `/groups/{id}/gifs`
Returns all gifs for grouping corresponding to the specified `{id}` parameter.
e.g. `curl http://localhost:1323/api/v1/groups/1/groups`

##### POST `/groups`
Creates a new gif grouping.
e.g. `curl -F "name=[group_name]" -F "image=@[image_path] http://localhost:1323/api/v1/groups`

##### POST `/groups/{id}/gifs`
Creates a new gif within grouping corresponding to the specified `{id}` parameter.
e.g. `curl -F "image=@[image_path] http://localhost:1323/api/v1/groups/{id}/gifs`
