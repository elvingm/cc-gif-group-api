package main

import (
    "net/http"
    "os"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo"
    mw "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/labstack/echo/middleware"

    "github.com/elvingm/cc-gifgroup-api/Godeps/_workspace/src/github.com/garyburd/redigo/redis"
)

type Group struct {
    Id   int    `json:"id"`
    Name string `json:"name"`
}

type ResponseTemplate struct {
    Content    interface{} `json:"content"`
    ErrorCode  int         `json:"error_code"`
    ErrorText  string      `json:"error_text"`
    StatusCode int         `json:"status_code"`
    StatusText string      `json:"status_text"`
    Success    bool        `json:"success"`
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
    res := ResponseTemplate{}
    groupSlice := []Group{{1, "Foo"}, {2, "Bar"}}

    res.Content = groupSlice
    return c.JSON(http.StatusOK, res)
}

func getGroupGifs(c *echo.Context) error {
    res := ResponseTemplate{}

    return c.JSON(http.StatusOK, res)
}

func createGroup(c *echo.Context) error {
    res := ResponseTemplate{}
    group := Group{}

    // save to Redis when created
    res.Content = group
    return c.JSON(http.StatusOK, res)
}

func createGifInGroup(c *echo.Context) error {
    res := ResponseTemplate{}

    return c.JSON(http.StatusOK, res)
}
