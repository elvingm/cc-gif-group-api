package main

import (
    "encoding/json"
    "net/http"
    "os"
    "strconv"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo"
    mw "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo/middleware"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/garyburd/redigo/redis"
)

type Group struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

type Groups []Group

type ResponseTemplate struct {
    Content    interface{} `json:"content"`
    ErrorCode  int         `json:"error_code"`
    ErrorText  string      `json:"error_text"`
    StatusCode int         `json:"status_code"`
    StatusText string      `json:"status_text"`
    Success    bool        `json:"success"`
}

var groupSeq = 1 // default

func init() { // connect to Redis on init, and get highest group ID
    os.Setenv("redisPort", ":6379")

    rC := RedisConnection()
    defer rC.Close()

    reply, err := redis.Int(rC.Do("GET", "id:groups"))
    if err != nil {
        _, err := rC.Do("SET", "id:groups", 1) // initialize if doesn't exist
        ErrorHandler(err)

        reply = 1
    }

    groupSeq = reply
}

func main() {
    os.Setenv("apiPort", ":1323")

    e := echo.New()

    e.Use(mw.Logger())
    e.Use(mw.Recover())

    // Routes
    e.Get("/groups", getAllGroups)
    e.Get("/groups/:id/gifs", getGroupGifs)
    e.Post("/groups", createGroup)
    e.Post("/groups/:id/gifs", createGifInGroup)

    e.Run(os.Getenv("apiPort"))
}

func getAllGroups(c *echo.Context) error {
    var groups Groups

    rC := RedisConnection()
    defer rC.Close()

    groupKeys, err := rC.Do("KEYS", "group:*")
    ErrorHandler(err)

    for _, k := range groupKeys.([]interface{}) {
        var group Group

        result, err := rC.Do("GET", k.([]byte))
        ErrorHandler(err) // TODO: handle error here or return error code?

        if err := json.Unmarshal(result.([]byte), &group); err != nil {
            ErrorHandler(err)
        }
        groups = append(groups, group)
    }

    res := ResponseTemplate{} // TODO: DRY out repetition of setting response values
    res.Content = groups
    res.ErrorCode = 0
    res.ErrorText = "No Error"
    res.StatusCode = http.StatusOK
    res.StatusText = "OK"
    res.Success = true
    return c.JSON(http.StatusOK, res)
}

func getGroupGifs(c *echo.Context) error {
    res := ResponseTemplate{}

    return c.JSON(http.StatusOK, res)
}

func createGroup(c *echo.Context) error {
    g := Group{groupSeq, c.Form("name")}

    rC := RedisConnection()
    defer rC.Close()

    gJson, err := json.Marshal(g)
    ErrorHandler(err)

    _, err = rC.Do("SET", "group:"+strconv.Itoa(g.Id), gJson)
    ErrorHandler(err)

    reply, err := redis.Int(rC.Do("INCR", "id:groups"))
    ErrorHandler(err)

    groupSeq = reply
    res := ResponseTemplate{}
    res.Content = g // returns group info that was saved - return total_count?
    res.ErrorCode = 0
    res.ErrorText = "No Error"
    res.StatusCode = http.StatusOK
    res.StatusText = "OK"
    res.Success = true
    return c.JSON(http.StatusOK, res)
}

func createGifInGroup(c *echo.Context) error {
    res := ResponseTemplate{}

    return c.JSON(http.StatusOK, res)
}

func RedisConnection() redis.Conn {
    c, err := redis.Dial("tcp", os.Getenv("redisPort"))
    ErrorHandler(err)
    return c
}

// Handler
func ErrorHandler(err error) {
    if err != nil {
        panic(err)
    }
}
