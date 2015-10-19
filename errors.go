package main

import "net/http"

func SetInternalServerError(res *ResponseTemplate, errorCode int, errorText string) {
    /* Error Codes: 
        1 - Redis Error saving groups/gifs
        2 - S3 Error uploading group image
        3 - Redis Error finding groups/gifs
    */

    res.Success = false
    res.StatusCode = http.StatusInternalServerError
    res.StatusText = http.StatusText(http.StatusInternalServerError)
    res.ErrorCode = errorCode 
    res.ErrorText = errorText
}

func SetBadRequestError(res *ResponseTemplate, errorCode int, errorText string) {
    /* Error Codes: 
        4 - Invalid form (missing image or other)
    */

    res.Success = false
    res.StatusCode = http.StatusBadRequest
    res.StatusText = http.StatusText(http.StatusBadRequest)
    res.ErrorCode = errorCode 
    res.ErrorText = errorText
}
