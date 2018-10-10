package main
import (
   "encoding/json"
   "net/http"
   "net/http/httptest"
   "testing"
   "fmt"
   "bytes"
   "github.com/gin-gonic/gin"
   "github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {
    gin.SetMode(gin.TestMode)

    router := gin.New()
    
    r := httptest.NewRequest("POST", "http://dataprovider:8081/ad", bytes.NewBuffer([]byte(`{"country":"FRA","device":"DESKTOP"}`)))
	w := httptest.NewRecorder()
    router.ServeHTTP(w, r)
    fmt.Printf("\n\nCheck response %+v\n", w)
    if w.Code != http.StatusOK { // Should be 200
		t.Fatalf("Expecting status to be 200 got %v", w.Code)
	}
}
func performRequest(r *gin.Engine, method, path string, data interface{}) *httptest.ResponseRecorder {


    
    j, _ := json.Marshal(data)

    //var jsonStr = []byte(`{"country":"FRA","device":"DESKTOP"}`)
    req1, err := http.NewRequest("POST", "http://dataprovider:8081"+path, bytes.NewBuffer(j))
    req1.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req1)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()
    fmt.Printf("\n\nFirst %+v\n\n", resp)



    
    fmt.Printf("\n\nPOST %s %+v\n","http://dataprovider:8081"+path, bytes.NewBuffer(j) )
    req, _ := http.NewRequest("POST", "http://dataprovider:8081"+path, bytes.NewBuffer(j))
    //req.Header.Set("X-Custom-Header", "myvalue")
    req.Header.Set("Content-Type", "application/json")


   //req, _ := http.NewRequest(method, path, bytes.NewBuffer(j))
   //req.Header.Set("Content-Type", "application/json")
   w := httptest.NewRecorder()
    r.ServeHTTP(w, req1)
    
   //r.ServeHTTP(w, req1)
   fmt.Printf("\n\nSomething %+v\n\n", w)
   return w
}

type ValidCampaign struct {
	PlacementID string
	Country string
	Device string
	ExpectedCampaignID string
}
func GetValidCampaigns()[]ValidCampaign {
	return []ValidCampaign {
		//ValidCampaign{"3946ca64ff78d93ca61090a437cbb6b3", "FRA", "DESKTOP", "9c0abe51c6e6655d81de2d044d4fb194"},
		//ValidCampaign{"3946ca64ff78d93ca61090a437cbb6b3", "BEL", "DESKTOP", "d0f631ca1ddba8db3bcfcb9e057cdc98"},
		//ValidCampaign{"3946ca64ff78d93ca61090a437cbb6b3", "FRA", "MOBILE", "9c0abe51c6e6655d81de2d044d4fb194"},
		//ValidCampaign{"f64551fcd6f07823cb87971cfb914464", "FRA", "DESKTOP", "d0f631ca1ddba8db3bcfcb9e057cdc98"},
		//ValidCampaign{"f64551fcd6f07823cb87971cfb914464", "BEL", "DESKTOP", "d0f631ca1ddba8db3bcfcb9e057cdc98"},
	}
}
func TestValidCampaign(t *testing.T) {

	for _,c := range GetValidCampaigns() {
		validCampaignResult(t, c)
	}
}

func validCampaignResult(t *testing.T, v ValidCampaign) {

   // Build our expected body
   body := gin.H{
      "country": v.Country,
      "device": v.Device,
   }
   // Grab our router
   router := SetupRouter()
   // Perform a GET request with that handler.
   w := performRequest(router, "POST", fmt.Sprintf("/ad?placement=%s", v.PlacementID), body)

   fmt.Printf("\n/ad?placement=%s\n%+v\n", v.PlacementID,body)
   // Assert we encoded correctly,
   // the request gives a 200
   assert.Equal(t, http.StatusOK, w.Code)
   // Convert the JSON response to a map
   var response map[string]string
   err := json.Unmarshal([]byte(w.Body.String()), &response)

   fmt.Printf("\n\nUne erreur est survenue :\n %s\n %+v\n\n", w.Body.String(), response)
   // Grab the value & whether or not it exists
   value, exists := response["campaign"]
   // Make some assertions on the correctness of the response.
   assert.Nil(t, err)
   assert.True(t, exists)
   assert.Equal(t, value, v.ExpectedCampaignID)
   
   _, existsContent := response["content"]
   assert.True(t, existsContent)

   // Grab the value & whether or not it exists
   _, errorExists := response["error"]
   assert.False(t, errorExists)
}





type InvalidCampaign struct {
	PlacementID string
	Country string
	Device string
	ExpectedError string
	ErrorCode int
}
func GetInvalidCampaigns()[]InvalidCampaign {
	return []InvalidCampaign {
		//InvalidCampaign{"f", "FRA", "DESKTOP", "unable to find a matching campaign for counrty FRA, device DESKTOP and placement f", http.StatusNotFound},
		//InvalidCampaign{"3946ca64ff78d93ca61090a437cbb6b3", "BEA", "DESKTOP", "Key: 'CampaignRequest.Country' Error:Field validation for 'Country' failed on the 'validCountry' tag", http.StatusBadRequest},
		//InvalidCampaign{"3946ca64ff78d93ca61090a437cbb6b3", "FRA", "TV", "Key: 'CampaignRequest.Device' Error:Field validation for 'Device' failed on the 'eq|eq|eq' tag", http.StatusBadRequest},
		//InvalidCampaign{"f64551fcd6f07823cb87971cfb914464", "FRA", "TABLET", "unable to find a matching campaign for counrty FRA, device TABLET and placement f64551fcd6f07823cb87971cfb914464", http.StatusNotFound},
	}
}
func TestInvalidCampaign(t *testing.T) {

	for _,c := range GetInvalidCampaigns() {
		invalidCampaignResult(t, c)
	}
}

func invalidCampaignResult(t *testing.T, v InvalidCampaign) {

   // Build our expected body
   body := gin.H{
      "country": v.Country,
      "device": v.Device,
   }
   // Grab our router
   router := SetupRouter()
   // Perform a GET request with that handler.
   w := performRequest(router, "POST", fmt.Sprintf("/ad?placement=%s", v.PlacementID), body)
   // Assert we encoded correctly,
   // the request gives a 200
   assert.Equal(t, v.ErrorCode, w.Code)
   // Convert the JSON response to a map
   var response map[string]string
   err := json.Unmarshal([]byte(w.Body.String()), &response)
   // Grab the value & whether or not it exists
   _, exists := response["campaign"]
   // Make some assertions on the correctness of the response.
   assert.Nil(t, err)
   assert.False(t, exists)
   
   _, existsContent := response["content"]
   assert.False(t, existsContent)

   // Grab the value & whether or not it exists
   errorValue, errorExists := response["error"]
   assert.True(t, errorExists)
   assert.Equal(t, errorValue, v.ExpectedError)
}

func SetupRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
	router := gin.New()	
	return router
 }
