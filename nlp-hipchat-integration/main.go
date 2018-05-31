package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"bitbucket.org/atlassianlabs/hipchat-golang-base/util"

	"bytes"
	"fmt"
	"io/ioutil"

	"regexp"

	"git.swisscom.ch/DOS/hipchat-go/hipchat"
	"github.com/gorilla/mux"
)

// RoomConfig holds information to send messages to a specific room
type RoomConfig struct {
	token *hipchat.OAuthAccessToken
	hc    *hipchat.Client
	name  string
}

// Context keep context of the running application
type Context struct {
	baseURL string
	static  string
	//rooms per room OAuth configuration and client
	rooms map[string]*RoomConfig
}

type HipchatMessage struct {
	Item Item `json:"item"`
}

type Item struct {
	Message MessageDetails `json:"message"`
	Room    Room           `json:"room"`
}

type MessageDetails struct {
	MessageText string `json:"message"`
	Sender      Sender `json:"from"`
}

type Room struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Sender struct {
	Id          int    `json:"id"`
	MentionName string `json:"mention_name"`
	Name        string `json:"name"`
}

func (c *Context) healthcheck(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode([]string{"OK"})
}

func (c *Context) atlassianConnect(w http.ResponseWriter, r *http.Request) {
	lp := path.Join("./static", "atlassian-connect.json")
	vals := map[string]string{
		"LocalBaseUrl": c.baseURL,
	}
	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "config", vals)
}

func (c *Context) installable(w http.ResponseWriter, r *http.Request) {
	authPayload, err := util.DecodePostJSON(r, true)
	if err != nil {
		log.Fatalf("Parsed auth data failed:%v\n", err)
	}

	credentials := hipchat.ClientCredentials{
		ClientID:     authPayload["oauthId"].(string),
		ClientSecret: authPayload["oauthSecret"].(string),
	}
	roomName := strconv.Itoa(int(authPayload["roomId"].(float64)))
	newClient := hipchat.NewClient("")
	tok, _, err := newClient.GenerateToken(credentials, []string{hipchat.ScopeSendNotification})
	if err != nil {
		log.Fatalf("Client.GetAccessToken returns an error %v", err)
	}
	rc := &RoomConfig{
		name: roomName,
		hc:   tok.CreateClient(),
	}
	c.rooms[roomName] = rc

	util.PrintDump(w, r, false)
	json.NewEncoder(w).Encode([]string{"OK"})
}

func (c *Context) config(w http.ResponseWriter, r *http.Request) {
	signedRequest := r.URL.Query().Get("signed_request")
	lp := path.Join("./static", "layout.hbs")
	fp := path.Join("./static", "config.hbs")
	vals := map[string]string{
		"LocalBaseUrl":  c.baseURL,
		"SignedRequest": signedRequest,
		"HostScriptUrl": c.baseURL,
	}
	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		log.Fatalf("%v", err)
	}
	tmpl.ExecuteTemplate(w, "layout", vals)
}

func insertMentions(message string, mentionsname string) string {
	mentionsPlaceholder := regexp.MustCompile("^(.*?)XXX(.*)$")
	repStr := "${1}" + "@" + mentionsname + "$2"
	return mentionsPlaceholder.ReplaceAllString(message, repStr)
}

func (c *Context) hook(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println("An error occurred reading request body. "+err.Error()+"\"", false)
		return
	}

	var hipchatmsg = &HipchatMessage{}
	err = json.Unmarshal(body, &hipchatmsg)
	if err != nil {
		fmt.Println("An error occurred parsing the body to json. "+err.Error()+"\"", false)
		return
	}
	roomID := strconv.Itoa(hipchatmsg.Item.Room.Id)
	util.PrintDump(w, r, true)

	nextAction := startConversation(hipchatmsg.Item.Message.MessageText, hipchatmsg.Item.Message.Sender.Id)
	fmt.Println(nextAction)
	nextAction = executeNextAction(nextAction, hipchatmsg.Item.Message.Sender.Id)
	fmt.Println(nextAction)
	if (nextAction == "action_listen") || (strings.Contains(nextAction, "utter_")) {
		textresponse := getResponseMessage(hipchatmsg.Item.Message.MessageText, hipchatmsg.Item.Message.Sender.Id, hipchatmsg.Item.Message.Sender.MentionName)

		fmt.Printf("Action response fron RASA is %v", textresponse)
		log.Printf("Sending notification to %d\n", hipchatmsg.Item.Room.Id)
		c.sendHipChatNotification(roomID, textresponse)
		return
	}
}

func startConversation(message string, senderid int) string {
	data := []byte("{\"query\": \"" + message + "\"}")
	responseRasa := SendRequestToRasa(data, strconv.Itoa(senderid), "/parse")

	rasabody, _ := ioutil.ReadAll(responseRasa.Body)
	fmt.Println(string(rasabody))
	var f interface{}
	err := json.Unmarshal(rasabody, &f)
	if err != nil {
		log.Printf("Unmarshaling error %v", err)
	}
	m := f.(map[string]interface{})

	nextAction := m["next_action"]
	return nextAction.(string)
}

func executeNextAction(action string, senderid int) string {
	data := []byte("{\"executed_action\": \"" + action + "\"}")
	responseRasa := SendRequestToRasa(data, strconv.Itoa(senderid), "/continue")
	rasabody, _ := ioutil.ReadAll(responseRasa.Body)
	fmt.Println(string(rasabody))
	var f interface{}
	err := json.Unmarshal(rasabody, &f)
	if err != nil {
		log.Printf("Unmarshaling error %v", err)
	}
	m := f.(map[string]interface{})

	nextAction := m["next_action"]
	return nextAction.(string)
}

func getResponseMessage(message string, senderid int, mentionsname string) string {
	data := []byte("{\"query\": \"" + message + "\"}")
	responseRasa := SendRequestToRasa(data, strconv.Itoa(senderid), "/respond")

	rasabody, _ := ioutil.ReadAll(responseRasa.Body)
	fmt.Println(string(rasabody))
	f := []interface{}{}
	err := json.Unmarshal(rasabody, &f)
	if err != nil {
		log.Printf("Unmarshaling error %s", err)
	}
	m2 := f[0].(map[string]interface{})

	textresponse := m2["text"].(string)
	// replace XXX with mentions
	textresponse = insertMentions(textresponse, mentionsname)
	return textresponse
}
func (c *Context) sendHipChatNotification(roomID string, message string) {
	log.Printf("Sending notification to %s\n", roomID)

	notifRq := &hipchat.NotificationRequest{
		Message:       string(message),
		MessageFormat: "html",
		Color:         "green",
	}

	if _, ok := c.rooms[roomID]; ok {
		_, err := c.rooms[roomID].hc.Room.Notification(roomID, notifRq)
		if err != nil {
			log.Printf("Failed to notify HipChat channel:%v\n", err)
		}
	} else {
		log.Printf("Room is not registered correctly:%v\n", c.rooms)
	}
}

func SendRequestToRasa(body []byte, userid string, method string) *http.Response {
	url := os.Getenv("CHATBOT_URL") + "/conversations/" + userid + method
	//url := "http://localhost:8080/conversations/" + userid + method

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		fmt.Printf("http.NewRequest() error: %v\n", err)
		return nil
	}

	c := &http.Client{}
	resp, err := c.Do(req)

	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return nil
	}
	return resp
}

// routes all URL routes for app add-on
func (c *Context) routes() *mux.Router {
	r := mux.NewRouter()
	//healthcheck route required by Micros
	r.Path("/healthcheck").Methods("GET").HandlerFunc(c.healthcheck)
	//descriptor for Atlassian Connect
	r.Path("/").Methods("GET").HandlerFunc(c.atlassianConnect)
	r.Path("/atlassian-connect.json").Methods("GET").HandlerFunc(c.atlassianConnect)

	// HipChat specific API routes
	r.Path("/installable").Methods("POST").HandlerFunc(c.installable)
	r.Path("/config").Methods("GET").HandlerFunc(c.config)
	r.Path("/hook").Methods("POST").HandlerFunc(c.hook)

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(c.static)))
	return r
}

func main() {
	var (
		port    = flag.String("port", "8080", "web server port")
		static  = flag.String("static", "./static/", "static folder")
		baseURL = flag.String("baseurl", os.Getenv("BASE_URL"), "local base url")
	)
	flag.Parse()

	c := &Context{
		baseURL: *baseURL,
		static:  *static,
		rooms:   make(map[string]*RoomConfig),
	}

	log.Printf("Base HipChat integration v0.10 - running on port:%v", *port)

	r := c.routes()
	http.Handle("/", r)
	http.ListenAndServe(":"+*port, nil)
}
