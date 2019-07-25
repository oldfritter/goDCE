package utils

import (
    "github.com/labstack/echo"
    "strings"
)

func GetRealIp(context echo.Context)(ip string) {
    ips := context.RealIP()
    //some forwarded_ips like 223.104.64.228,183.240.52.39, 172.68.254.56, sduppid operators
    rips := strings.Split(ips, ",")
    ip = rips[0]
    return
}