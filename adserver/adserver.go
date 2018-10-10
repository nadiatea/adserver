package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
	"reflect"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopkg.in/go-playground/validator.v8"
	"io/ioutil"
	"encoding/json"
	"github.com/Shopify/sarama"
	
)
const TOPIC_CONSUME_CAMPAIGN = "consume.campaign"

type CampaignRequest struct {
	Country string `json:"country" binding:"required,validCountry"`
	Device string `json:"device" binding:"required,eq=DESKTOP|eq=TABLET|eq=MOBILE"`
}

type CampaignStore struct {
	Campaigns map[string]CampaignItem `json:"campaigns"`

}

type CampaignItem struct {
	Price float64 `json:"price"`
	Content CampaignContent `json:"content"`
	Countries []string `json:"countries"`
	Devices []string `json:"devices"`
	Placements []string `json:"placements"`

}

type Campaign struct {
	CampaignID string `json:"campaign"`
	Content CampaignContent `json:"content"`
}

type CampaignContent struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Landing string `json:"landing"`
}

var campaigns CampaignStore
var producer sarama.SyncProducer

func main() {
	var campaignsFile string

	campaignsFile = "/campaigns.json"

	if os.Getenv("CAMPAIGNS_FILE") != "" {
		campaignsFile = os.Getenv("CAMPAIGNS_FILE")
	}
	port := ":8081"
	//argsWithProg := os.Args
	argsWithoutProg := os.Args[1:]
	for i:=0; i<len(argsWithoutProg); i++ {
		if argsWithoutProg[i] == "-p" {
			port = ":" + argsWithoutProg[i+1]
		} else if argsWithoutProg[i] == "-f" {
			campaignsFile = argsWithoutProg[i+1]
		}
		i++
	}

	if _, err := os.Stat(campaignsFile); os.IsNotExist(err) {
		panic(fmt.Errorf("Please specify a campaign file"))
	}

	connectToKafka()

	cacheCampaigns(campaignsFile)

	router := gin.Default()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validCountry", validateCountry)
	}


	// Listen campaign requests
	router.POST("/ad", func(c *gin.Context) {
		placementID := c.Query("placement")
		var request CampaignRequest
		// Check posting data
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Find the best campaign
		bestCampaign, price, err := fetchCampaign(request, placementID)
		if err != nil {
			// No mathing campaign
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		postLog(placementID, bestCampaign.CampaignID, price)
		c.JSON(http.StatusOK, bestCampaign)
	})
	router.Run(port) // listen and serve on 0.0.0.0:8081
}

// cacheCampaigns set the campaigns in memory
func cacheCampaigns(campaignsFile string) {
	campaignsJson, err := ioutil.ReadFile(campaignsFile)
	if err != nil {
        panic(err)
	}

	raw := []byte(campaignsJson)

	json.Unmarshal(raw, &campaigns)

}

// fetchCampaign fetch the matching campaign with the best price
func fetchCampaign(request CampaignRequest, emplacementID string) (bestCampaign *Campaign, bestPrice float64, err error) {
	for campaignID, campaign := range campaigns.Campaigns {
		if len(campaign.Countries)>0 && !contains(campaign.Countries, request.Country) {
			continue
		} 
		if !contains(campaign.Devices, request.Device) {
			continue
		}
		if !contains(campaign.Placements, emplacementID) {
			continue
		}
		if campaign.Price > bestPrice {
			bestPrice = campaign.Price
			bestCampaign = &Campaign{CampaignID:campaignID, Content:campaign.Content}
		}
	}
	if bestCampaign != nil {
		return
	}
	return nil, bestPrice, fmt.Errorf("unable to find a matching campaign for counrty %s, device %s and placement %s", request.Country, request.Device, emplacementID)
}



// contains is utility function to find if a string is in a string array
func contains(s []string, e string) bool {
    for _, a := range s {
        if a == e {
            return true
        }
    }
    return false
}

var countries = []string{"AND", "ARE", "AFG", "ATG", "AIA", "ALB", "ARM", "AGO", "ATA", "ARG", "ASM", "AUT", "AUS", "ABW", "ALA", "AZE", "BIH", "BRB", "BGD", "BEL", "BFA", "BGR", "BHR", "BDI", "BEN", "BLM", "BMU", "BRN", "BOL", "BES", "BRA", "BHS", "BTN", "BVT", "BWA", "BLR", "BLZ", "CAN", "CCK", "COD", "CAF", "COG", "CHE", "CIV", "COK", "CHL", "CMR", "CHN", "COL", "CRI", "CUB", "CPV", "CUW", "CXR", "CYP", "CZE", "DEU", "DJI", "DNK", "DMA", "DOM", "DZA", "ECU", "EST", "EGY", "ESH", "ERI", "ESP", "ETH", "FIN", "FJI", "FLK", "FSM", "FRO", "FRA", "GAB", "GBR", "GRD", "GEO", "GUF", "GGY", "GHA", "GIB", "GRL", "GMB", "GIN", "GLP", "GNQ", "GRC", "SGS", "GTM", "GUM", "GNB", "GUY", "HKG", "HMD", "HND", "HRV", "HTI", "HUN", "IDN", "IRL", "ISR", "IMN", "IND", "IOT", "IRQ", "IRN", "ISL", "ITA", "JEY", "JAM", "JOR", "JPN", "KEN", "KGZ", "KHM", "KIR", "COM", "KNA", "PRK", "KOR", "XKX", "KWT", "CYM", "KAZ", "LAO", "LBN", "LCA", "LIE", "LKA", "LBR", "LSO", "LTU", "LUX", "LVA", "LBY", "MAR", "MCO", "MDA", "MNE", "MAF", "MDG", "MHL", "MKD", "MLI", "MMR", "MNG", "MAC", "MNP", "MTQ", "MRT", "MSR", "MLT", "MUS", "MDV", "MWI", "MEX", "MYS", "MOZ", "NAM", "NCL", "NER", "NFK", "NGA", "NIC", "NLD", "NOR", "NPL", "NRU", "NIU", "NZL", "OMN", "PAN", "PER", "PYF", "PNG", "PHL", "PAK", "POL", "SPM", "PCN", "PRI", "PSE", "PRT", "PLW", "PRY", "QAT", "REU", "ROU", "SRB", "RUS", "RWA", "SAU", "SLB", "SYC", "SDN", "SSD", "SWE", "SGP", "SHN", "SVN", "SJM", "SVK", "SLE", "SMR", "SEN", "SOM", "SUR", "STP", "SLV", "SXM", "SYR", "SWZ", "TCA", "TCD", "ATF", "TGO", "THA", "TJK", "TKL", "TLS", "TKM", "TUN", "TON", "TUR", "TTO", "TUV", "TWN", "TZA", "UKR", "UGA", "UMI", "USA", "URY", "UZB", "VAT", "VCT", "VEN", "VGB", "VIR", "VNM", "VUT", "WLF", "WSM", "YEM", "MYT", "ZAF", "ZMB", "ZWE", "SCG", "ANT"}

func validateCountry(
	v *validator.Validate, topStruct reflect.Value, currentStructOrField reflect.Value,
	field reflect.Value, fieldType reflect.Type, fieldKind reflect.Kind, param string,
) bool {
	if country, ok := field.Interface().(string); !ok || !contains(countries, country) {
		return false
	}
	return true
}


func connectToKafka() (err error) {
	messageBrokers := "localhost:9092"
	if os.Getenv("MESSAGE_BROKERS") != "" {
		messageBrokers = os.Getenv("MESSAGE_BROKERS")
	}

	//addresses of available kafka brokers
	brokers := []string{messageBrokers}
	//setup relevant config info
	config := sarama.NewConfig()
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	producer, err = sarama.NewSyncProducer(brokers, config)
	if err != nil {
		panic(err)
	}
	defer producer.Close()
	return
}

func sendAMessage(topic string, msg string) error {
	//topic := "kafka-topic-for-the-message" //e.g create-user-topic
	var partition int32 = -1 //Partition to produce to 
	//msg := "actual information to save on kafka" //e.g {"name":"John Doe", "email":"john.doe@email.com"}
	message := &sarama.ProducerMessage{
					 Topic: topic,
					 Partition: partition,
					 Value: sarama.StringEncoder(msg), 
			   }
	//partition, offset, err := producer.SendMessage(message)
	
	_, _, err := producer.SendMessage(message)
	return err
	
}

func postLog(placementID string, campaignID string, price float64) error {
	return sendAMessage(TOPIC_CONSUME_CAMPAIGN, fmt.Sprintf(`{"timestamp":%d, "placementId":"%s", "campaignId":"%s", "price":%0.f}`, int32(time.Now().Unix()), placementID, campaignID, price))
}