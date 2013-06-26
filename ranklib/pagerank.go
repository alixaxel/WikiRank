package ranklib

import (
  "math"
  "log"
)

func pageRank(pages []Page, walkProbability float64, convergenceCriteron float64) ([]float64, map[uint64] uint32) {
  beta, epsilon := walkProbability, convergenceCriteron
  log.Printf("Ranking with beta='%f', epsilon='%f'", beta, epsilon)
  n := len(pages)
  idRemap := make(map[uint64] uint32, n)
  lastRank := make([]float64, n)
  thisRank := make([]float64, n)

  for i := 0; i < n; i++ {
    idRemap[pages[i].Id] = uint32(i)
  }

  for iteration, lastChange := 1, math.MaxFloat64; lastChange > epsilon; iteration++ {
    thisRank, lastRank = lastRank, thisRank
    if iteration > 1 {
      // Clear out old values
      for i:=0; i < n; i++ {
        thisRank[i] = 0.0
      }
    } else {
      // Base case: everything uniform
      for i:= 0; i < n; i++ {
        lastRank[i] = 1.0 / float64(n)
      }
    }

    // Single power iteration
    for i := 0; i < n; i++ {
      contribution := beta * lastRank[i] / float64(len(pages[i].Links))
      for _, link := range pages[i].Links {
        thisRank[idRemap[link.PageId]] += contribution
      }
    }

    // Reinsert leaked probability
    S := float64(0.0)
    for i := 0; i < n; i++ {
      S += thisRank[i]
    }
    leakedRank := (1.0 - S) / float64(n)
    lastChange = 0.0 // and calculate L1-difference too
    for i := 0; i < n; i++ {
      thisRank[i] += leakedRank
      lastChange += math.Abs(thisRank[i] - lastRank[i])
    }

    log.Printf("Pagerank iteration #%d delta=%f", iteration, lastChange)
  }

  return thisRank, idRemap
}
