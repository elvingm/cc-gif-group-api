package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "strconv"

    "github.com/labstack/echo"
    mw "github.com/labstack/echo/middleware"

    "github.com/garyburd/redigo/redis"
    "github.com/mitchellh/goamz/aws"
    "github.com/mitchellh/goamz/s3"
    "github.com/tmilewski/goenv"
)

type Group struct {
    Id       int    `json:"id"`
    Name     string `json:"name"`
    ImageUrl string `json:"image_url"`
}

type Gif struct {
    Id       int    `json:"id"`
    GroupId  int    `json:"group_id"`
    ImageUrl string `json:"image_url"`
}

type Groups []Group
type Gifs []Gif

type ResponseTemplate struct {
    Content    interface{} `json:"content"`
    ErrorCode  int         `json:"error_code"`
    ErrorText  string      `json:"error_text"`
    StatusCode int         `json:"status_code"`
    StatusText string      `json:"status_text"`
    Success    bool        `json:"success"`
}

var groupSeq = 1 // default
var gifSeq = 1   // default

func init() { // connect to Redis on init, and get highest group ID
    err := goenv.Load()
    if err != nil {
        fmt.Println("Missing environment variables file")
        panic(err)
    }

    rC := RedisConnection()
    defer rC.Close()

    groupSeqValue, err := redis.Int(rC.Do("GET", "id:groups"))
    if err != nil {
        _, err := rC.Do("SET", "id:groups", 1) // initialize if doesn't exist
        ErrorHandler(err)

        groupSeqValue = 1
    }

    gifSeqValue, err := redis.Int(rC.Do("GET", "id:gifs"))
    if err != nil {
        _, err := rC.Do("SET", "id:gifs", 1) // initialize if doesn't exist
        ErrorHandler(err)

        gifSeqValue = 1
    }

    groupSeq = groupSeqValue
    gifSeq = gifSeqValue
}

func main() {
    e := echo.New()

    e.Use(mw.Logger())
    e.Use(mw.Recover())

    // Routes
    e.Get("/groups", GetGroups)
    e.Get("/groups/:id/gifs", GetGroupGifs)
    e.Post("/groups", PostGroups)
    e.Post("/groups/:id/gifs", PostGroupGif)

    e.Run(os.Getenv("API_PORT"))
}

// Route Functions
func GetGroups(c *echo.Context) error {
    res := NewResponseTemplate()

    groups, err := FindAllGroups(res)
    if err != nil {
        SetInternalServerError(res, 3, "Server error finding groups")
        return c.JSON(res.StatusCode, res)
    }

    res.Content = groups
    return c.JSON(res.StatusCode, res)
}

func GetGroupGifs(c *echo.Context) error {
    res := NewResponseTemplate()
    groupId, err := strconv.Atoi(c.Param("id"))
    ErrorHandler(err)

    gifs, err := FindGroupGifs(c.Request(), groupId)
    if err != nil {
        SetInternalServerError(res, 3, "Server error finding group gifs")
    }

    res.Content = gifs

    return c.JSON(http.StatusOK, res)
}

func PostGroups(c *echo.Context) error {
    res := NewResponseTemplate()

    group := &Group{}
    group.Id = groupSeq
    if group.Name = "Unnamed Group"; len(c.Form("name")) > 0 {
        group.Name = c.Form("name")
    }

    err := SaveGroupImage(c.Request(), group, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    err = SaveGroup(group, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    res.Content = group
    return c.JSON(http.StatusOK, res)
}

func PostGroupGif(c *echo.Context) error {
    res := NewResponseTemplate()
    groupId, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        SetBadRequestError(res, 4, "Invalid group id for gif")
    }

    gif := &Gif{}
    gif.Id = gifSeq
    gif.GroupId = groupId

    err = SaveGifToGroup(c.Request(), gif, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    err = SaveGif(gif, res)
    if err != nil {
        return c.JSON(res.StatusCode, res)
    }

    res.Content = gif
    return c.JSON(http.StatusOK, res)
}

// Util Functions
func RedisConnection() redis.Conn {
    c, err := redis.Dial("tcp", os.Getenv("REDIS_PORT"))
    ErrorHandler(err)
    return c
}

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

func NewResponseTemplate() *ResponseTemplate {
    template := &ResponseTemplate{}
    template.Success = true
    template.StatusCode = http.StatusOK
    template.StatusText = http.StatusText(http.StatusOK)
    template.ErrorCode = 0
    template.ErrorText = "No Error"
    return template
}

// DB Access Functions
func FindAllGroups(res *ResponseTemplate) (Groups, error) {
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

    return groups, nil
}

func FindGroupGifs(res *http.Request, groupId int) (Gifs, error) {
    var gifs Gifs
    rC := RedisConnection()
    defer rC.Close()

    gifKeys, err := rC.Do("SMEMBERS", "gifsForGroup:"+strconv.Itoa(groupId))
    ErrorHandler(err)

    for _, k := range gifKeys.([]interface{}) {
        var gif Gif

        result, err := rC.Do("GET", k.([]byte))
        ErrorHandler(err) // TODO: handle error here?

        if err := json.Unmarshal(result.([]byte), &gif); err != nil {
            ErrorHandler(err)
        }

        gifs = append(gifs, gif)
    }

    return gifs, nil
}

func SaveGroup(g *Group, res *ResponseTemplate) error {
    rC := RedisConnection()
    defer rC.Close()

    gJson, err := json.Marshal(g)
    ErrorHandler(err)

    _, err = rC.Do("SET", "group:"+strconv.Itoa(g.Id), gJson)
    if err != nil {
        SetInternalServerError(res, 1, "Error saving group")
        return err
    }

    reply, err := redis.Int(rC.Do("INCR", "id:groups"))
    if err != nil {
        SetInternalServerError(res, 1, "Error incrementing id:groups count")
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
        SetBadRequestError(res, 4, "Invalid or missing image")
        return err
    }

    path := fmt.Sprintf("groups/%v/%v", g.Id, header.Filename)

    err = bucket.Put(path, content, req.Header.Get("Content-Type"), s3.PublicRead)
    if err != nil {
        SetInternalServerError(res, 2, "Error uploading image to bucket")
        return err
    }

    g.ImageUrl = bucket.URL(path)
    return nil
}

func SaveGif(gif *Gif, res *ResponseTemplate) error {
    rC := RedisConnection()
    defer rC.Close()

    gifJson, err := json.Marshal(gif)
    ErrorHandler(err)

    _, err = rC.Do("SET", "gif:"+strconv.Itoa(gif.Id), gifJson)
    if err != nil {
        SetInternalServerError(res, 1, "Error saving group")
        return err
    }

    _, err = rC.Do("SADD", "gifsForGroup:"+strconv.Itoa(gif.GroupId), "gif:"+strconv.Itoa(gif.Id))
    reply, err := redis.Int(rC.Do("INCR", "id:gifs"))
    if err != nil {
        SetInternalServerError(res, 1, "Error incrementing id:groups count")
        return err
    }

    gifSeq = reply
    return nil
}

func SaveGifToGroup(req *http.Request, g *Gif, res *ResponseTemplate) error {
    bucket := S3Bucket()
    req.ParseMultipartForm(16 << 20)

    image, header, err := req.FormFile("image")
    if err != nil {
        SetBadRequestError(res, 4, "Invalid or missing image")
        return err
    }

    content, err := ioutil.ReadAll(image)
    if err != nil {
        SetBadRequestError(res, 4, "Invalid or missing image")
        return err
    }

    path := fmt.Sprintf("groups/%v/gifs/%v", g.GroupId, header.Filename)

    err = bucket.Put(path, content, req.Header.Get("Content-Type"), s3.PublicRead)
    if err != nil {
        SetInternalServerError(res, 2, "Error uploading image to bucket")
        return err
    }

    g.ImageUrl = bucket.URL(path)

    return nil
}
