package service

import (
  "container/heap"
  "encoding/json"
  "fmt"
  "github.com/pkg/errors"
  "io"
  "math"
  "net/http"
)

const earthRadiusKm = 6371 // Earth's radius in kilometers

type ReverseLocation struct {
  DisplayName string `json:"display_name"`
  Address     struct {
    HouseNumber string `json:"house_number"`
    Road        string `json:"road"`
    Suburb      string `json:"suburb"`
    Country     string `json:"county"`
    City        string `json:"city"`
    State       string `json:"state"`
  } `json:"address"`
}

// Graph represents an adjacency list graph
type Graph struct {
  Nodes []*Node
}

// Node represents a graph node with latitude and longitude
type Node struct {
  ID       int
  Lat, Lon float64
  Adjacent map[*Node]float64
  Dist     float64
  Index    int // The index of the node in the heap.
}

// PriorityQueue implements heap.Interface and holds Nodes.
type PriorityQueue []*Node

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
  return pq[i].Dist < pq[j].Dist
}

func (pq PriorityQueue) Swap(i, j int) {
  pq[i], pq[j] = pq[j], pq[i]
  pq[i].Index = i
  pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
  n := len(*pq)
  node := x.(*Node)
  node.Index = n
  *pq = append(*pq, node)
}

func (pq *PriorityQueue) Pop() interface{} {
  old := *pq
  n := len(old)
  node := old[n-1]
  old[n-1] = nil  // avoid memory leak
  node.Index = -1 // for safety
  *pq = old[0 : n-1]
  return node
}

// NewNode creates a new Node with given latitude and longitude
func NewNode(id int, lat, lon float64) *Node {
  return &Node{ID: id, Lat: lat, Lon: lon, Adjacent: make(map[*Node]float64), Dist: math.MaxInt32}
}

// AddEdge adds an edge to the Node
func (n *Node) AddEdge(neighbor *Node) {
  n.Adjacent[neighbor] = haversine(n.Lat, n.Lon, neighbor.Lat, neighbor.Lon)
}

// haversine calculates the great-circle distance between two points on a sphere
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
  dLat := (lat2 - lat1) * math.Pi / 180.0
  dLon := (lon2 - lon1) * math.Pi / 180.0

  a := math.Sin(dLat/2)*math.Sin(dLat/2) +
    math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
      math.Sin(dLon/2)*math.Sin(dLon/2)

  c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
  distance := earthRadiusKm * c * 1000

  return distance // Distance in kilometers
}

// Dijkstra calculates shortest path
func Dijkstra(graph *Graph, start *Node) {
  pq := make(PriorityQueue, 1)
  pq[0] = start
  start.Dist = 0
  heap.Init(&pq)

  for pq.Len() > 0 {
    u := heap.Pop(&pq).(*Node)

    for v, weight := range u.Adjacent {
      if dist := u.Dist + weight; v.Dist > dist {
        v.Dist = dist
        heap.Push(&pq, v)
      }
    }
  }
}

func FindBranchesWithinRadius(branches []*Node, clientLat, clientLon, radius float64) []*Node {
  var driversWithinRadius []*Node

  for _, driver := range branches {
    distance := haversine(clientLat, clientLon, driver.Lat, driver.Lon)
    if distance <= radius {
      driversWithinRadius = append(driversWithinRadius, driver)
    }
  }

  return driversWithinRadius
}

//func NearPoint() {
//  graph := &Graph{Nodes: make([]*Node, 5)}
//
//  // Initialize nodes with latitude and longitude
//  // Example coordinates: Replace these with actual latitudes and longitudes
//
//  graph.Nodes[0] = NewNode(0, 41.361237, 69.2082421) // Example for 0 my location
//  graph.Nodes[1] = NewNode(1, 41.346102, 69.2038551) // Example for 1 beruniy
//  graph.Nodes[2] = NewNode(2, 41.333084, 69.2163801) // Example for 2 tinchlik
//
//  // Initialize more nodes as needed
//
//  // Add edges
//  graph.Nodes[0].AddEdge(graph.Nodes[1])
//  graph.Nodes[0].AddEdge(graph.Nodes[2])
//  // Add more edges as needed
//
//  // Run Dijkstra's algorithm
//  Dijkstra(graph, graph.Nodes[0])
//
//  // Print shortest distances
//  for _, node := range graph.Nodes {
//    fmt.Printf("Shortest distance from node 0 to node %d is %d km\n", node.ID, node.Dist)
//  }
//}


func CalculateDistance(la1, lo1, la2, lo2 float64) float64 {
	lat1 := la1 * math.Pi / 180.0
	lon1 := lo1 * math.Pi / 180.0
	lat2 := la2 * math.Pi / 180.0
	lon2 := lo2 * math.Pi / 180.0
  
	dLon := lon2 - lon1
	dLat := lat2 - lat1
  
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1)*math.Cos(lat2)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
  
	// Masofa (meter)
	distance := earthRadiusKm * c * 1000
	return distance
  }
  
  func Reverse(lat, lon float64) (ReverseLocation, error) {
	client := http.Client{}
	endpoint := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse")
  
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
	  //c.JSON(http.StatusOK, gin.H{
	  //  "error":  fmt.Sprintf("client: could not create request: %s", err.Error()),
	  //  "status": false,
	  //})
  
	  return ReverseLocation{}, errors.New(fmt.Sprintf("client: could not create request: %s", err.Error()))
	}
  
	values := req.URL.Query()
	values.Add("format", "json")
	values.Add("lat", fmt.Sprintf("%f", lat))
	values.Add("lon", fmt.Sprintf("%f", lon))
	req.URL.RawQuery = values.Encode()
  
	res, err := client.Do(req)
	if err != nil {
	  //c.JSON(http.StatusOK, gin.H{
	  //  "error":  fmt.Sprintf("client: error making http request: %s", err.Error()),
	  //  "status": false,
	  //})
  
	  return ReverseLocation{}, errors.New(fmt.Sprintf("client: error making http request: %s", err.Error()))
	}
  
	defer func(Body io.ReadCloser) {
	  err = Body.Close()
	  if err != nil {
		//c.JSON(http.StatusOK, gin.H{
		//  "error":  err.Error(),
		//  "status": false,
		//})
		return
	  }
	}(res.Body)
  
	data, err := io.ReadAll(res.Body)
	if err != nil {
	  //c.JSON(http.StatusOK, gin.H{
	  //  "error":  err.Error(),
	  //  "status": false,
	  //})
  
	  return ReverseLocation{}, err
	}
  
	if res.StatusCode != http.StatusOK {
	  //c.JSON(http.StatusOK, gin.H{
	  //  "error":  "wrong request!",
	  //  "status": false,
	  //})
  
	  return ReverseLocation{}, errors.New("wrong request!")
	}
  
	var result ReverseLocation
	err = json.Unmarshal(data, &result)
	if err != nil {
	  //c.JSON(http.StatusOK, gin.H{
	  //  "error":  err.Error(),
	  //  "status": false,
	  //})
  
	  return ReverseLocation{}, err
	}
  
	return result, nil
  }
  
