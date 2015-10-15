package main

import (
    "fmt"
    "encoding/json"
    "io/ioutil"
    "net/http"
    "os"
    "strconv"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo"
    mw "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo/middleware"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/garyburd/redigo/redis"
    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/mitchellh/goamz/aws"
    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/mitchellh/goamz/s3"
)

type Group struct {
    Id            int    `json:"id"`
    Name          string `json:"name"`
    ImageUrl      string `json:"group_image_url"`
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
    os.Setenv("AWS_ACCESS_KEY_ID", "abc") // placeholders until replaced with env management
    os.Setenv("AWS_SECRET_ACCESS_KEY", "123")

    e := echo.New()

    e.Use(mw.Logger())
    e.Use(mw.Recover())

    // Routes
    e.Get("/groups", GetGroups)
    e.Get("/groups/:id/gifs", getGroupGifs)
    e.Post("/groups", PostGroups)
    e.Post("/groups/:id/gifs", createGifInGroup)

    e.Run(os.Getenv("apiPort"))
}

func GetGroups(c *echo.Context) error {
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

func PostGroups(c *echo.Context) error {
    res := &ResponseTemplate{}
    res.Success = true
    res.StatusCode = http.StatusOK
    res.ErrorCode = 0
    res.StatusText = http.StatusText(http.StatusOK)
    res.ErrorCode = "No Error"
    
    g := &Group{}
    g.Id = groupSeq
    if g.Name = "Unnamed Group"; len(c.Form("name")) > 0 {
        g.Name = c.Form("name")
    }

    err := SaveGroupImage(c.Request(), g, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    err = SaveGroup(g, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    res.Content = g
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

func S3Bucket() *s3.Bucket {
    auth, err := aws.EnvAuth()
    ErrorHandler(err)

    client := s3.New(auth, aws.USEast)
    b := client.Bucket("cc-gifgroup-api")
    return b
}

func SaveGroup(g *Group, res *ResponseTemplate) error {
    rC := RedisConnection()
    defer rC.Close()

    gJson, err := json.Marshal(g)
    ErrorHandler(err)

    _, err = rC.Do("SET", "group:"+strconv.Itoa(g.Id), gJson)
    if err != nil {
        res.Success = false
        res.StatusCode = http.StatusInternalServerError
        res.StatusText = http.StatusText(http.StatusInternalServerError)
        res.ErrorCode = 1 // app-specific, 1 = redis error
        res.ErrorText = "Error saving group"

        return err
    }

    reply, err := redis.Int(rC.Do("INCR", "id:groups"))
    if err != nil {
        res.Success = false
        res.StatusCode = http.StatusInternalServerError
        res.StatusText = http.StatusText(http.StatusInternalServerError)
        res.ErrorCode = 1 // app-specific error code, 1 = image upload error
        res.ErrorText = "Error incrementing id:groups count"

        return err
    }

    groupSeq = reply
    return nil
}

func SaveGroupImage(req *http.Request, g *Group, res *ResponseTemplate) error {
    bucket := S3Bucket()
    req.ParseMultipartForm(16 << 20)

    image, header, err := req.FormFile("image")
    if err != nil {
        g.ImageUrl = bucket.URL("default/group-default.gif")
        return nil
    }

    content, err := ioutil.ReadAll(image)
    if err != nil {
        res.Success = false
        res.StatusCode = http.StatusBadRequest
        res.StatusText = http.StatusText(http.StatusBadRequest)

        return err
    }

    path := fmt.Sprintf("groups/%v/%v", g.Id, header.Filename)

    err = bucket.Put(path, content, req.Header.Get("Content-Type"), s3.PublicRead)
    if err != nil {
        res.Success = false
        res.StatusCode = http.StatusInternalServerError
        res.StatusText = http.StatusText(http.StatusInternalServerError)
        res.ErrorCode = 2 // app-specific error code, 2 = s3 image upload error
        res.ErrorText = "Error uploading image to bucket"

        return err
    }
    
    g.ImageUrl = bucket.URL(path)
    return nil
}