package main

import (
	"fmt"
	"os"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/Shopify/sarama"
)

/*
Bonus​: un second serveur s’occupera
a. d’écrire dans un fichier texte chaque affichage de pub réussie d’une campagne sur
un placement où chaque ligne aura le format suivant:
timestamp, placementId, campaignId, price
b. de logger les mêmes informations dans la sortie standard.
c. de répondre à une requête GET sur un handler nommé “/sum” permettant de
recevoir la somme des prix des affichages de pub réussie depuis le lancement de
l’adserver.
d. de répondre à une requête GET sur un handler nommé “/sum_placement” la somme
des prix, par campagne, des affichages de pub réussies sur un emplacement donné
en paramètre de “query”.
Voici un exemple d’une requête GET:
http://localhost:8080/sum_placement?placement=f64551fcd6f07823cb87971cfb914464
Voici un exemple de réponse:
{
"d0f631ca1ddba8db3bcfcb9e057cdc98":10,
"9c0abe51c6e6655d81de2d044d4fb194":6.9
}
*/

var logPath string
var logFile *os.File

var sum float64
var sumByCampaign = map[string]map[string]float64{}

const TOPIC_CONSUME_CAMPAIGN = "consume.campaign"


func main() {
	port := ":8081"
	logPath = "adDisplayPath.txt"

	//argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]
	for i:=0; i<len(argsWithoutProg); i++ {
		if argsWithoutProg[i] == "-p" {
			//TODO check port format
			port = ":" + argsWithoutProg[i+1]
		} else if argsWithoutProg[i] == "-f" {
			//TODO check file format
			logPath = argsWithoutProg[i+1]
		}
		i++
	}

	logFile, err := os.Create(logPath)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	router := gin.Default()

	// Listen sum by campaign requests
	router.POST("/sum", func(c *gin.Context) {
		/*
		// Find the best campaign
		result, err := fetchSum()
		if err != nil {
			// No mathing campaign
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
		*/
		c.JSON(http.StatusOK, sum)
	})

	// Listen sum by campaign requests
	router.POST("/sum_placement", func(c *gin.Context) {
		placementID := c.Query("placement")

		// Check posting data
		//TODO check query params correctly
		if placementID != "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Please furnish a placement"})
			return
		}

		// Find the best campaign
		/*
		result, err := fetchSumByCampaign(placementID)
		if err != nil {
			// No mathing campaign
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, result)
		*/
		if sum, exists := sumByCampaign[placementID]; exists {
			c.JSON(http.StatusOK, sum)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "unable to find the placement ID"})
		}
	})
	router.Run(port) // listen and serve on 0.0.0.0:8081
}

// writeLog write in a log the display of an ad
func writeLog(timestamp int, placementID string, campaignID string, price float64) {
	log := fmt.Sprintf("%d, %s, %s, %0.f\n", timestamp, placementID, campaignID, price)
	_, err := logFile.WriteString(log)
	if err!= nil {
		panic(err)
	}
	logFile.Sync()
	fmt.Print(log)
}

// fetchSum fetch the sum of prices of all display of ad
func fetchSum() (sum float64, err error) {
	return 
}

// fetchSumByCampaign fetch the sum of prices of all display of ad by campaign
func fetchSumByCampaign(emplacementID string) (result map[string]float64, err error) {
	return 
}

func consumeKafkaMessages() {
	messageBrokers := "localhost:9092"
	if os.Getenv("MESSAGE_BROKERS") != "" {
		messageBrokers = os.Getenv("MESSAGE_BROKERS")
	}
	//addresses of available kafka brokers
	brokers := []string{messageBrokers}
	consumer, err := sarama.NewConsumer(brokers, nil)

	topic := TOPIC_CONSUME_CAMPAIGN //e.g. user-created-topic
	partitionList, err := consumer.Partitions(topic) //get all partitions
	if err != nil {
		panic(err)
	}
	//messages := make(chan *sarama.ConsumerMessage, 256)
	initialOffset := sarama.OffsetOldest //offset to start reading message from
	for _, partition := range partitionList {  
		pc, _ := consumer.ConsumePartition(topic, partition, initialOffset)
		go func(pc sarama.PartitionConsumer) {
			for message := range pc.Messages() {
				//messages <- message //or call a function that writes to disk

				data := struct {
					Timestamp int `json:"timestamp"`
					PlacementID string `json:"placementId"`
					CampaignID string `json:"campaignId"`
					Price float64 `json:"price"`
				}{}
				
				if err := json.Unmarshal(message.Value, &data); err != nil {
					panic(err)
				}
				writeLog(data.Timestamp, data.PlacementID, data.CampaignID, data.Price)
				if s, exists := sumByCampaign[data.PlacementID]; exists {
					if _, exists2 := s[data.CampaignID];exists2 {
						sumByCampaign[data.PlacementID][data.CampaignID] = data.Price
					} else {
						sumByCampaign[data.PlacementID][data.CampaignID] += data.Price
					}
				} else {
					sumByCampaign[data.PlacementID] = map[string]float64{data.CampaignID:data.Price}
				}
			}
		}(pc)
	}
}