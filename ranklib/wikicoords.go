package ranklib

import (
  "fmt"
  "log"
  "regexp"
  "strings"
  "strconv"
)


type Coordinate struct {
  lat float64
  long float64
}

func mustCompileInfobox(key string) *regexp.Regexp {
  return regexp.MustCompile(fmt.Sprintf(`(?i)[|] *%s *= ([^|]*)`, key))
}

func extractFromInfobox(wikiText string, r *regexp.Regexp) string {
  s := r.FindStringSubmatch(wikiText)
  if s != nil {
    return strings.TrimSpace(s[len(s) - 1])
  } else {
    return ""
  }
}

var latDR = mustCompileInfobox("(latd|lat_d|lat_degrees|latdegrees)")
var latMR = mustCompileInfobox("(latm|lat_m|lat_minutes|latminutes)")
var latSR = mustCompileInfobox("(lats|lat_s|lat_seconds|latseconds)")
var latNSR = mustCompileInfobox("(latNS|lat_NS|lat_direction|latdirection)")

var longDR = mustCompileInfobox("(longd|long_d|long_degrees|longdegrees)")
var longMR = mustCompileInfobox("(longm|long_m|long_minutes|longminutes)")
var longSR = mustCompileInfobox("(longs|long_s|long_seconds|longseconds)")
var longEWR = mustCompileInfobox("(longEW|long_EW|long_direction|longdirection)")

var coordInfoboxR = regexp.MustCompile(fmt.Sprintf(`(?i)[|] *(coordinates|coord|coordinate) *= ({{.*?}})`))

func optionalFloat(s string) (float64, error) {
  if s == "" {
    return 0, nil
  } else {
    return strconv.ParseFloat(s, 64)
  }
}

func coordinateFromStrings(
  latdString string, latmString string, latsString string, latNS string,
  longdString string, longmString string, longsString string, longEW string) (Coordinate, bool) {
    latd, err := strconv.ParseFloat(strings.TrimSpace(latdString), 64)
    latm, err := optionalFloat(strings.TrimSpace(latmString))
    lats, err := optionalFloat(strings.TrimSpace(latsString))

    longd, err := strconv.ParseFloat(strings.TrimSpace(longdString), 64)
    longm, err := optionalFloat(strings.TrimSpace(longmString))
    longs, err := optionalFloat(strings.TrimSpace(longsString))

    if err != nil {
      return Coordinate{}, false
    }

    latd = latd + (latm * 60 + lats) / 3600
    longd = longd + (longm * 60 + longs)  / 3600

    if strings.HasPrefix(latNS,  "S") {
      latd *= -1
    }
    if strings.HasPrefix(longEW, "W") {
      longd *= -1
    }
    return Coordinate{lat: latd, long: longd}, true
}

func coordinateFromInfobox(pe *pageElement) (Coordinate, bool) {
  coordString := extractFromInfobox(pe.Text, coordInfoboxR)
  if coordString != "" {
    c, ok := decimalCoordinate(coordString)
    if ok {
      return c, true
    } else {
      log.Printf("Weird coordinate in '%s' infobox - %s", pe.Title, coordString)
    }
  }

  latdString := extractFromInfobox(pe.Text, latDR)
  latmString := extractFromInfobox(pe.Text, latMR)
  latsString := extractFromInfobox(pe.Text, latSR)
  latNS := extractFromInfobox(pe.Text, latNSR)

  longdString := extractFromInfobox(pe.Text, longDR)
  longmString := extractFromInfobox(pe.Text, longMR)
  longsString := extractFromInfobox(pe.Text, longSR)
  longEW := extractFromInfobox(pe.Text, longEWR)

  if(latdString == "" || longdString == "") {
    return Coordinate{}, false
  }
  return coordinateFromStrings(latdString, latmString, latsString, latNS, longdString, longmString, longsString, longEW)
}

var middleBit = regexp.MustCompile("(?i){{coord *[|](.*)?}}")
func decimalCoordinate(wikiCoord string) (Coordinate, bool) {
  // See http://en.wikipedia.org/wiki/Template:Coord/doc/internals
  middle := middleBit.FindStringSubmatch(wikiCoord)
  if middle == nil {
    log.Printf("Error coordinate format: '%s'", wikiCoord)
    return Coordinate{}, false
  }
  params := strings.Split(middle[1], "|")

  start := -1
  end := len(params)
  for i, s := range params {
    s := strings.TrimSpace(s)
    _, err := strconv.ParseFloat(s, 64)
    if err == nil && start == -1 {
      start = i
      continue
    }

    if start >= 0 && err != nil && s != "N" && s != "E" && s != "S" && s != "W" {
      end = i
      break
    }
  }

  fmtLength := end - start
  var latdString, latmString, latsString, latNS string
  var longdString, longmString, longsString, longEW string
  if fmtLength == 2 {
    // dec
    latdString = params[start + 0]
    longdString = params[start + 1]
  } else if fmtLength == 4 {
    // d
    latdString = params[start + 0]
    latNS = params[start + 1]
    longdString = params[start + 2]
    longEW = params[start + 3]
  } else if fmtLength == 6 {
    // dm
    latdString = params[start + 0]
    latmString = params[start + 1]
    latNS = params[start + 2]
    longdString = params[start + 3]
    longmString = params[start + 4]
    longEW = params[start + 5]
  } else if fmtLength == 8 {
    //dms
    latdString = params[start + 0]
    latmString = params[start + 1]
    latsString = params[start + 2]
    latNS = params[start + 3]
    longdString = params[start + 4]
    longmString = params[start + 5]
    longsString = params[start + 6]
    longEW = params[start + 7]
  } else {
    log.Printf("Error coordinate format: '%s' (start=%d)", wikiCoord, start)
    return Coordinate{}, false
  }

  return coordinateFromStrings(latdString, latmString, latsString, latNS, longdString, longmString, longsString, longEW)
}

var linkRegex = regexp.MustCompile(`\[\[(?:([^|\]]*)\|)?([^\]]+)\]\]`)
var cleanSectionRegex = regexp.MustCompile(`^[^#]*`)
var coordinateRegex = regexp.MustCompile(`(?i){{coord(.*?)}}`)

func coordinatesFromWikiText(pe *pageElement) (Coordinate, bool) {
  if coord, ok := coordinateFromInfobox(pe); ok {
    return coord, ok
  }

  coordinateText := coordinateRegex.FindString(pe.Text)
  if coordinateText != "" && strings.Contains(coordinateText, "title") && !strings.Contains(coordinateText, "LAT") {
    coord, ok := decimalCoordinate(coordinateText)
    if ok {
      return coord, ok
    } else {
      log.Printf("Error was in '%s'", pe.Title)
    }
  }

  return Coordinate{}, false
}
