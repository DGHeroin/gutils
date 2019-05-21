package gutils

import "fmt"

func BytesToString(s uint64) string {
    unit := ""
    if s > 1024 {
        s = s / 1024
    }

    if s > 1024 {
        s = s / 1024
        unit = "K"
    }

    if s > 1024 {
        s = s / 1024
        unit = "M"
    }

    if s > 1024 {
        s = s / 1024
        unit = "G"
    }

    if s > 1024 {
        s = s / 1024
        unit = "T"
    }

    return fmt.Sprintf("%v%v", s, unit)
}