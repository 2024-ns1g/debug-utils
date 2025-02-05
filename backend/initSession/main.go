package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ANSI escape codes for colors
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// 各種Payload
type UserRegistrationPayload struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type RoomCreationPayload struct {
	DisplayName string `json:"displayName"`
}
type SlideCreationPayload struct {
	DisplayName string `json:"displayName"`
	Summary     string `json:"summary"`
}
type PageCreationPayload struct {
	Content string `json:"content"`
}
type ScriptCreationPayload struct {
	ScriptContent string `json:"scriptContent"`
}
type VoteTemplateCreationPayload struct {
	SlideId  string `json:"slideId"`
	Index    int    `json:"index,omitempty"`
	Title    string `json:"title"`
	Question string `json:"question"`
}
type VoteOptionCreationPayload struct {
	TemplateId      string `json:"templateId"`
	Index           int    `json:"index,omitempty"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	BorderColor     string `json:"borderColor,omitempty"`
}
type SessionCreationPayload struct {
	SessionId string `json:"sessionId"`
}
type OtpIssuePayload struct {
	Otp string `json:"otp"`
}

// VoteTemplateInfo - 投票テンプレートと、その下にあるオプションのIDをまとめて保持
type VoteTemplateInfo struct {
	ID        string
	Title     string
	Question  string
	OptionIDs []string
}

// HTTPPost sends a POST request with optional Bearer token
func HTTPPost(url string, payload interface{}, token string) ([]byte, error) {
	var request *http.Request
	var reqErr error

	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
		if req != nil {
			request = req
		}
		if err != nil {
			reqErr = err
		}
	} else {
		req, err := http.NewRequest("POST", url, nil)
		if req != nil {
			request = req
		}
		if err != nil {
			reqErr = err
		}
	}

	if reqErr != nil {
		return nil, fmt.Errorf("failed to create request: %w", reqErr)
	}

	request.Header.Set("Content-Type", "application/json")
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("received error status code: %d", resp.StatusCode)
	}

	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	return responseData, nil
}

// printTable prints a table with colored borders and headers
func printTable(headers []string, rows [][]string) {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print top border
	fmt.Print(colorCyan + "+")
	for _, w := range widths {
		fmt.Print("-" + strings.Repeat("-", w) + "-+")
	}
	fmt.Println(colorReset)

	// Print header row
	fmt.Print(colorCyan + "|")
	for i, header := range headers {
		fmt.Printf(" %-*s |", widths[i], header)
	}
	fmt.Println(colorReset)

	// Print separator
	fmt.Print(colorCyan + "+")
	for _, w := range widths {
		fmt.Print("-" + strings.Repeat("-", w) + "-+")
	}
	fmt.Println(colorReset)

	// Print all data rows
	for _, row := range rows {
		fmt.Print(colorCyan + "|")
		for i, cell := range row {
			fmt.Printf(" %-*s |", widths[i], cell)
		}
		fmt.Println(colorReset)
	}

	// Print bottom border
	fmt.Print(colorCyan + "+")
	for _, w := range widths {
		fmt.Print("-" + strings.Repeat("-", w) + "-+")
	}
	fmt.Println(colorReset)
}

func main() {
	// コマンドラインフラグの設定（新たに baseUrl フラグを追加）
	baseUrlFlag := flag.String("baseUrl", "http://localhost:8080", "Base URL for the API endpoints")
	username := flag.String("username", "test", "Username for registration")
	password := flag.String("password", "test", "Password for registration")
	roomName := flag.String("room", "test-room", "Name of the room")
	slideName := flag.String("slide", "test-slide", "Name of the slide")
	slideSummary := flag.String("summary", "This is a test slide", "Summary of the slide")
	scriptContent := flag.String("script", "This is a test script", "Content of the script")

	flag.Parse()

	// API のエンドポイントを baseUrl から生成
	baseURL := *baseUrlFlag
	registerURL := baseURL + "/auth/username/register"
	createRoomURL := baseURL + "/room/create"
	createSlideURLTemplate := baseURL + "/room/%s/slide/create"
	createPageURLTemplate := baseURL + "/room/%s/slide/%s/page/create"
	createScriptURLTemplate := baseURL + "/room/%s/slide/%s/page/%s/script/create"
	createVoteTemplateURLTemplate := baseURL + "/room/%s/slide/%s/vote/create"
	createVoteOptionURLTemplate := baseURL + "/room/%s/slide/%s/vote/%s/option/create"
	startSessionURL := baseURL + "/room/%s/slide/%s/session/create"
	issueAgentOtpURL := baseURL + "/room/%s/slide/%s/session/%s/agent/issue"
	issueAudienceOtpURL := baseURL + "/room/%s/slide/%s/session/%s/audience/issue"
	issuePresenterOtpURL := baseURL + "/room/%s/slide/%s/session/%s/presenter/issue"

	fmt.Println(colorGreen + "Initializing resources..." + colorReset)

	// Step 1: User registration
	fmt.Println(colorYellow + "Registering user..." + colorReset)
	userPayload := UserRegistrationPayload{
		Username: *username,
		Password: *password,
	}
	userResponse, err := HTTPPost(registerURL, userPayload, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error registering user: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var userResponseData map[string]string
	json.Unmarshal(userResponse, &userResponseData)
	token := userResponseData["token"]

	// Step 2: Room creation
	fmt.Println(colorYellow + "Creating room..." + colorReset)
	roomPayload := RoomCreationPayload{DisplayName: *roomName}
	roomResponse, err := HTTPPost(createRoomURL, roomPayload, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error creating room: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var roomData map[string]string
	json.Unmarshal(roomResponse, &roomData)
	roomID := roomData["roomId"]

	// Step 3: Slide creation
	fmt.Println(colorYellow + "Creating slide..." + colorReset)
	slideURL := fmt.Sprintf(createSlideURLTemplate, roomID)
	slidePayload := SlideCreationPayload{
		DisplayName: *slideName,
		Summary:     *slideSummary,
	}
	slideResponse, err := HTTPPost(slideURL, slidePayload, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error creating slide: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var slideData map[string]string
	json.Unmarshal(slideResponse, &slideData)
	slideID := slideData["slideId"]

	// Step 4: Create multiple pages with Markdown content
	pageContents := []string{
		"# Page 1\n\nThis is the first page with some **Markdown** content.\n\n- Item 1\n- Item 2\n- Item 3",
		"# Page 2\n\nThis is the second page with more **Markdown** content.\n\n- Item A\n- Item B\n- Item C",
		"# Page 3\n\nThis is the third page with even more **Markdown** content.\n\n- Item X\n- Item Y\n- Item Z",
	}

	pageIDs := make([]string, 0, len(pageContents))
	for i, content := range pageContents {
		fmt.Printf(colorYellow+"Creating page %d...\n"+colorReset, i+1)
		pageURL := fmt.Sprintf(createPageURLTemplate, roomID, slideID)
		pagePayload := PageCreationPayload{Content: content}
		pageResponse, err := HTTPPost(pageURL, pagePayload, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, colorRed+"Error creating page %d: %v\n"+colorReset, i+1, err)
			os.Exit(1)
		}
		var pageData map[string]string
		json.Unmarshal(pageResponse, &pageData)
		pageID := pageData["pageId"]
		pageIDs = append(pageIDs, pageID)
	}

	// Step 5: Create scripts for each page
	for i, pageID := range pageIDs {
		fmt.Printf(colorYellow+"Creating script for page %d...\n"+colorReset, i+1)
		scriptURL := fmt.Sprintf(createScriptURLTemplate, roomID, slideID, pageID)
		scriptPayload := ScriptCreationPayload{ScriptContent: *scriptContent}
		_, err := HTTPPost(scriptURL, scriptPayload, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, colorRed+"Error creating script for page %d: %v\n"+colorReset, i+1, err)
			os.Exit(1)
		}
	}

	// Step 6: Create multiple vote templates and options
	voteTitles := []string{"Vote 1", "Vote 2", "Vote 3"}
	voteQuestions := []string{"Is this the first vote?", "Is this the second vote?", "Is this the third vote?"}
	voteOptionTitles := []string{"Option A", "Option B", "Option C"}
	voteOptionDescriptions := []string{"First option", "Second option", "Third option"}

	var createdVoteTemplates []VoteTemplateInfo

	for i := range pageIDs {
		fmt.Printf(colorYellow+"Creating vote template for page %d...\n"+colorReset, i+1)
		voteTemplateURL := fmt.Sprintf(createVoteTemplateURLTemplate, roomID, slideID)
		voteTemplatePayload := VoteTemplateCreationPayload{
			SlideId:  slideID,
			Title:    voteTitles[i],
			Question: voteQuestions[i],
		}
		voteTemplateResponse, err := HTTPPost(voteTemplateURL, voteTemplatePayload, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, colorRed+"Error creating vote template for page %d: %v\n"+colorReset, i+1, err)
			os.Exit(1)
		}
		var voteTemplateData map[string]string
		json.Unmarshal(voteTemplateResponse, &voteTemplateData)
		voteTemplateID := voteTemplateData["id"] // CUIDなどのIDが入る想定

		// VoteTemplateInfo を用意
		tempInfo := VoteTemplateInfo{
			ID:       voteTemplateID,
			Title:    voteTitles[i],
			Question: voteQuestions[i],
		}

		// Create vote options for the vote template
		for j := range voteOptionTitles {
			fmt.Printf(colorYellow+"Creating vote option %d for vote template %d...\n"+colorReset, j+1, i+1)
			voteOptionURL := fmt.Sprintf(createVoteOptionURLTemplate, roomID, slideID, voteTemplateID)
			voteOptionPayload := VoteOptionCreationPayload{
				TemplateId:  voteTemplateID,
				Title:       voteOptionTitles[j],
				Description: voteOptionDescriptions[j],
			}
			voteOptionResponse, err := HTTPPost(voteOptionURL, voteOptionPayload, token)
			if err != nil {
				fmt.Fprintf(os.Stderr, colorRed+"Error creating vote option %d for vote template %d: %v\n"+colorReset, j+1, i+1, err)
				os.Exit(1)
			}
			var voteOptionData map[string]string
			json.Unmarshal(voteOptionResponse, &voteOptionData)
			optionID := voteOptionData["id"] // CUIDなどのIDが入る想定

			tempInfo.OptionIDs = append(tempInfo.OptionIDs, optionID)
		}

		createdVoteTemplates = append(createdVoteTemplates, tempInfo)
	}

	// Step 7: Start session
	fmt.Println(colorYellow + "Starting session..." + colorReset)
	sessionURL := fmt.Sprintf(startSessionURL, roomID, slideID)
	sessionPayload := SessionCreationPayload{SessionId: "test-session"}
	sessionResponse, err := HTTPPost(sessionURL, sessionPayload, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error starting session: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var sessionData map[string]string
	json.Unmarshal(sessionResponse, &sessionData)
	sessionID := sessionData["sessionId"]

	// Step 8: Issue agent OTP
	fmt.Println(colorYellow + "Issuing agent OTP..." + colorReset)
	time.Sleep(2 * time.Second)
	agentOtpURL := fmt.Sprintf(issueAgentOtpURL, roomID, slideID, sessionID)
	agentOtpResponse, err := HTTPPost(agentOtpURL, nil, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error issuing agent OTP: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var agentOtpData map[string]string
	json.Unmarshal(agentOtpResponse, &agentOtpData)
	agentOtp := agentOtpData["otp"]

	// Step 9: Issue audience OTP
	fmt.Println(colorYellow + "Issuing audience OTP..." + colorReset)
	time.Sleep(2 * time.Second)
	audienceOtpURL := fmt.Sprintf(issueAudienceOtpURL, roomID, slideID, sessionID)
	audienceOtpResponse, err := HTTPPost(audienceOtpURL, nil, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error issuing audience OTP: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var audienceOtpData map[string]string
	json.Unmarshal(audienceOtpResponse, &audienceOtpData)
	audienceOtp := audienceOtpData["otp"]

	// Step 10: Issue presenter OTP
	fmt.Println(colorYellow + "Issuing presenter OTP..." + colorReset)
	time.Sleep(2 * time.Second)
	presenterOtpURL := fmt.Sprintf(issuePresenterOtpURL, roomID, slideID, sessionID)
	presenterOtpResponse, err := HTTPPost(presenterOtpURL, nil, token)
	if err != nil {
		fmt.Fprintf(os.Stderr, colorRed+"Error issuing presenter OTP: %v\n"+colorReset, err)
		os.Exit(1)
	}
	var presenterOtpData map[string]string
	json.Unmarshal(presenterOtpResponse, &presenterOtpData)
	presenterOtp := presenterOtpData["otp"]

	// ==========================================================
	// 最後にすべてまとめたサマリーテーブルを出力
	// ==========================================================

	headers := []string{"Resource", "ID", "Details"}
	var rows [][]string

	// ユーザ登録
	rows = append(rows, []string{
		"User",
		*username,
		"Registered successfully",
	})
	// ルーム
	rows = append(rows, []string{
		"Room",
		roomID,
		*roomName,
	})
	// スライド
	rows = append(rows, []string{
		"Slide",
		slideID,
		*slideName,
	})

	// VoteTemplate と Option の表示
	for i, vt := range createdVoteTemplates {
		// Template
		rows = append(rows, []string{
			fmt.Sprintf("Vote Template %d", i+1),
			vt.ID,
			fmt.Sprintf("Title: %s / Q: %s", vt.Title, vt.Question),
		})
		// Options
		for j, optID := range vt.OptionIDs {
			rows = append(rows, []string{
				fmt.Sprintf("Vote Option %d-%d", i+1, j+1),
				optID,
				fmt.Sprintf("Linked to Template %d", i+1),
			})
		}
	}

	// セッション
	rows = append(rows, []string{
		"Session",
		sessionID,
		"Started successfully",
	})
	// Agent OTP
	rows = append(rows, []string{
		"Agent OTP",
		agentOtp,
		"Issued successfully",
	})
	// Audience OTP
	rows = append(rows, []string{
		"Audience OTP",
		audienceOtp,
		"Issued successfully",
	})
	// Presenter OTP
	rows = append(rows, []string{
		"Presenter OTP",
		presenterOtp,
		"Issued successfully",
	})

	printTable(headers, rows)
	fmt.Println(colorGreen + "All resources created and OTPs issued successfully." + colorReset)
}
