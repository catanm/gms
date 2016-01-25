package main

import (
	"fmt"
	"strconv"
)

const (
	THRESH = 9 // Hamming distance to be considered same cluster
)

type Image struct {
	URL   string
	Phash string
}

/** Compute hamming distance **/
func distance(s1 string, s2 string) int {
	if s1 == "" || s2 == "" {
		return 32
	}
	var counter int = 0
	for k := 0; k < len(s1); k++ {
		if string(s1[k]) != string(s2[k]) {
			counter++
		}

	}
	return counter
}

/* Cluster images using single pass cluster for max thresh */
func cluster(imgs map[string]Image) []Cluster {
	var clusters []Cluster
	i := 0
	for _, cur := range imgs {
		if !cur.PassSize || !cur.PassSynthetic {
			continue
		}

		if len(clusters) == 0 {
			var imgs []Image
			clustImgs := append(imgs, cur)
			curCluster := Cluster{Images: clustImgs, Phash: cur.Phash}
			clusters = append(clusters, curCluster)
		} else {
			/* Check if thresh is small enough to merge, otherwise create new cluster */
			var added bool = false
			for j, clust := range clusters {
				dist := distance(cur.Phash, clust.Phash)
				//		fmt.Println("Distance: " + strconv.Itoa(dist))
				if dist < THRESH {
					/* Add image to list */
					clusters[j].Images = append(clust.Images, cur)
					added = true
					/* TODO: Update Phash */
					break
				}
			}

			/* Still not added? Create a new cluster */
			if !added {
				var imgs []Image
				clustImgs := append(imgs, cur)
				curCluster := Cluster{Images: clustImgs, Phash: cur.Phash}
				clusters = append(clusters, curCluster)
			}

		}
		i++
	}

	//	printClusters(clusters)
	return clusters

}

func printClusters(clusters []Cluster) {
	for k, clust := range clusters {
		fmt.Println("Cluster " + strconv.Itoa(k) + ": " + clust.Phash + " Images: " + strconv.Itoa(len(clust.Images)))
		for _, img := range clust.Images {
			fmt.Println("-- " + img.URL)
		}
	}
}
