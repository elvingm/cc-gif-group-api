package main

import(
  "github.com/labstack/echo"
  mw "github.com/labstack/echo/middleware"
)

func main() {
  e := echo.New()
  e.Use(mw.Logger())
  e.Use(mw.Recover())

  e.Static("/", "public")
  e.Run(":1323")
}