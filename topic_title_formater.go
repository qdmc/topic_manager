package topic_manager

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var targetReg, matchReg, isMatchReg, titleMatch *regexp.Regexp
var singleMatchStr, multipleMatchStr string

func init() {
	singleMatchStr = `[0-9a-zA-Z_\-]+`
	multipleMatchStr = fmt.Sprintf(`([0-9a-zA-Z_\-]+%s)*[0-9a-zA-Z_\-]+`, LayerSeparation)
	targetReg = regexp.MustCompile(`^[0-9a-zA-Z_\-]+$`)
	titleMatch = regexp.MustCompile(fmt.Sprintf(`^%s([0-9a-zA-Z_\-]+%s)*[0-9a-zA-Z_\-]+$`, LayerSeparation, LayerSeparation))
	//matchReg = regexp.MustCompile(`(\+*/*)*#+/*(\+*/*)*`)
	matchReg = regexp.MustCompile(fmt.Sprintf(`(\+*%s*)*#+%s*(\+*%s*)*`, LayerSeparation, LayerSeparation, LayerSeparation))
	isMatchReg = regexp.MustCompile(`[#|+]`)
}

type titleFormat struct {
	title   string
	isMatch bool
	targets []string
	match   *regexp.Regexp
}

// checkPublishTopicTitle    校验发布时的topic
func checkPublishTopicTitle(topicTitle string) (string, error) {
	var title string
	if title == "" {
		return title, errors.New("topicTitle is empty")
	}
	if strings.Index(topicTitle, LayerSeparation) != 0 {
		title = LayerSeparation + title
	}
	if !titleMatch.MatchString(title) {
		return title, errors.New("topicTitle is not regexp")
	}
	return title, nil
}

// formatTitle          格式化title,在topic的新增,订阅时调用
func formatTitle(topicTitle string) (TopicInterface, error) {
	isMatch, title, targets, err := checkTopicTitle(topicTitle)
	if err != nil {
		return nil, err
	}
	f := &titleFormat{
		title:   title,
		isMatch: isMatch,
		targets: targets,
		match:   nil,
	}
	if !isMatch {
		return newCommonTopic(f), nil
	}
	if title == "/#" {
		f.match = regexp.MustCompile(fmt.Sprintf(`^/%s$`, multipleMatchStr))
		return newCommonTopic(f), nil
	} else if title == "/+" {
		f.match = regexp.MustCompile(fmt.Sprintf(`^/%s$`, singleMatchStr))
		return newCommonTopic(f), nil
	}
	regStr := `^`
	for index, target := range targets {
		if target == "#" {
			regStr += fmt.Sprintf("/%s", multipleMatchStr)
		} else if target == "+" {
			regStr += fmt.Sprintf("/%s", singleMatchStr)
		} else {
			regStr += fmt.Sprintf("/%s", target)
		}
		if index == len(targets)-1 {
			regStr += "$"
		}
	}
	match, err := regexp.Compile(regStr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("reg.Compile.Err(%s): %s", regStr, topicTitle))
	}
	f.match = match
	return newCommonTopic(f), nil
}

// checkTarget        校验标识[0-9a-zA-Z_-]
func checkTarget(t string) error {
	if t == "#" || t == "+" {
		return nil
	}
	if !targetReg.MatchString(t) {
		return errors.New("target regex is error")
	}
	return nil
}

// checkTopicTitle    校验title并分组
func checkTopicTitle(topicTitle string) (isMatch bool, title string, targets []string, err error) {
	// 这个是根目录,返回错误
	if topicTitle == "/" {
		err = errors.New(`"/" is root topic`)
		return
	} else if topicTitle == "#" || topicTitle == "/#" {
		isMatch = true
		title = "/#"
		targets = []string{"#"}
		return
	} else if topicTitle == "+" || topicTitle == "/+" {
		isMatch = true
		title = "/+"
		targets = []string{"+"}
		return
	}
	if topicTitle == "" {
		err = errors.New("topicTitle is empty")
		return
	}
	isMatch = isMatchReg.MatchString(topicTitle)
	if isMatch {
		// "#" 只能出现一次,todo :"#"多次出现后的快速匹配
		if strings.Count(topicTitle, "#") > 1 {
			err = errors.New(`"#" must be only once`)
			return
		}
		// 处理'#'与'+'相连
		topicTitle = matchReg.ReplaceAllString(topicTitle, "#")
		// 处理'#'与'+'前后的分隔
		topicTitle = strings.ReplaceAll(topicTitle, "#", "/#/")
		topicTitle = strings.ReplaceAll(topicTitle, "+", "/+/")
	}
	// 去前后空格
	topicTitle = strings.Trim(topicTitle, " ")
	// 去前后的分隔符
	topicTitle = strings.Trim(topicTitle, LayerSeparation)
	if topicTitle == "" {
		err = errors.New("trim topicTitle is empty")
		return
	}

	spArr := strings.Split(topicTitle, LayerSeparation)
	for _, target := range spArr {
		if target == "" {
			continue
		}
		err = checkTarget(target)
		if err != nil {
			err = errors.New(fmt.Sprintf(`%s --> %s`, target, err.Error()))
			return
		} else {
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		err = errors.New("split topicTitle is empty")
		return
	}
	title = fmt.Sprintf("%s%s", LayerSeparation, strings.Join(targets, LayerSeparation))
	// 最长254 字节
	if len([]byte(title)) > 254 {
		err = errors.New("topicTitle  is too long")
		return
	}
	return isMatch, title, targets, nil
}
