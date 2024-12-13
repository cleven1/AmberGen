package agent

import (
	"math"
	"strings"
	"unicode"
)

// 专业领域权重定义
var domainWeights = map[string]float64{
	"product_management":   0.9,  // 产品管理
	"ui_ux_design":         0.85, // UI/UX设计
	"ai_development":       0.95, // AI开发
	"data_science":         0.9,  // 数据科学
	"operation_management": 0.8,  // 运营管理
	"business_analysis":    0.85, // 商业分析
	"system_architecture":  0.95, // 系统架构
	"security":             0.9,  // 安全
	"testing":              0.8,  // 测试
}

// TF-IDF 相关参数
const (
	minWordLength = 3     // 最小词长度
	maxWordLength = 20    // 最大词长度
	minTermFreq   = 0.001 // 最小词频
	maxTermFreq   = 0.8   // 最大词频
)

// calculateAgentCapability 计算Agent的能力分数
func calculateAgentCapability(agent Agent) float64 {
	// 1. 获取基础权重
	baseWeight := getBaseDomainWeight(agent.GetCapabilities())

	// 2. 计算经验系数 (基于历史表现)
	experienceScore := calculateExperienceScore(agent)

	// 3. 计算专业度系数
	expertiseScore := calculateExpertiseScore(agent)

	// 4. 计算最终分数
	finalScore := baseWeight*0.4 + experienceScore*0.3 + expertiseScore*0.3

	// 5. 归一化处理 (确保分数在0.1-1.0之间)
	return normalizeScore(finalScore)
}

// calculateRelevance 计算输出与输入的相关性分数
func calculateRelevance(input, output string) float64 {
	// 1. 文本预处理
	inputTerms := preprocessText(input)
	outputTerms := preprocessText(output)

	// 2. 计算TF-IDF向量
	inputVector := calculateTFIDF(inputTerms)
	outputVector := calculateTFIDF(outputTerms)

	// 3. 计算余弦相似度
	similarity := cosineSimilarity(inputVector, outputVector)

	// 4. 计算长度比例分数
	lengthScore := calculateLengthScore(input, output)

	// 5. 计算结构相似度
	structureScore := calculateStructureScore(input, output)

	// 6. 综合评分
	finalScore := similarity*0.5 + lengthScore*0.2 + structureScore*0.3

	return normalizeScore(finalScore)
}

// getBaseDomainWeight 获取领域基础权重
func getBaseDomainWeight(capabilities []string) float64 {
	if len(capabilities) == 0 {
		return 0.5
	}

	totalWeight := 0.0
	for _, capability := range capabilities {
		if weight, exists := domainWeights[capability]; exists {
			totalWeight += weight
		} else {
			totalWeight += 0.5 // 默认权重
		}
	}

	return totalWeight / float64(len(capabilities))
}

// calculateExperienceScore 计算经验分数
func calculateExperienceScore(agent Agent) float64 {
	if expert, ok := agent.(*ExpertAgent); ok {
		// 基于历史记录计算成功率
		history := expert.memory.GetHistory()
		if len(history) == 0 {
			return 0.5
		}

		successfulResponses := 0
		for _, msg := range history {
			if msg.Role == "assistant" && isQualityResponse(msg.Content) {
				successfulResponses++
			}
		}

		return float64(successfulResponses) / float64(len(history))
	}
	return 0.5
}

// calculateExpertiseScore 计算专业度分数
func calculateExpertiseScore(agent Agent) float64 {
	if expert, ok := agent.(*ExpertAgent); ok {
		// 分析描述中的专业术语密度
		description := expert.description
		terms := extractProfessionalTerms(description)
		return float64(len(terms)) / float64(len(strings.Fields(description)))
	}
	return 0.5
}

// preprocessText 文本预处理
func preprocessText(text string) map[string]int {
	terms := make(map[string]int)
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		word = cleanWord(word)
		if isValidTerm(word) {
			terms[word]++
		}
	}

	return terms
}

// calculateTFIDF 计算TF-IDF向量
func calculateTFIDF(terms map[string]int) map[string]float64 {
	vector := make(map[string]float64)
	totalTerms := 0

	for _, count := range terms {
		totalTerms += count
	}

	for term, count := range terms {
		tf := float64(count) / float64(totalTerms)
		if tf >= minTermFreq && tf <= maxTermFreq {
			// 简化的IDF计算
			idf := math.Log(1.0 + 1.0/tf)
			vector[term] = tf * idf
		}
	}

	return vector
}

// cosineSimilarity 计算余弦相似度
func cosineSimilarity(v1, v2 map[string]float64) float64 {
	dotProduct := 0.0
	norm1 := 0.0
	norm2 := 0.0

	// 计算点积和向量范数
	for term, value1 := range v1 {
		if value2, exists := v2[term]; exists {
			dotProduct += value1 * value2
		}
		norm1 += value1 * value1
	}

	for _, value2 := range v2 {
		norm2 += value2 * value2
	}

	// 避免除零
	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// calculateLengthScore 计算长度比例分数
func calculateLengthScore(input, output string) float64 {
	inputLen := float64(len(strings.Fields(input)))
	outputLen := float64(len(strings.Fields(output)))

	// 理想的输出长度是输入的1-3倍
	ratio := outputLen / inputLen
	if ratio < 1 {
		return ratio
	}
	if ratio > 3 {
		return 3 / ratio
	}
	return 1.0
}

// calculateStructureScore 计算结构相似度
func calculateStructureScore(input, output string) float64 {
	// 分析段落结构
	inputParagraphs := len(strings.Split(input, "\n\n"))
	outputParagraphs := len(strings.Split(output, "\n\n"))

	// 分析标点符号使用
	inputPunct := countPunctuation(input)
	outputPunct := countPunctuation(output)

	// 计算段落比例
	paraScore := 1.0 - math.Abs(float64(outputParagraphs-inputParagraphs))/float64(inputParagraphs)

	// 计算标点符号使用的相似度
	punctScore := 1.0 - math.Abs(float64(outputPunct-inputPunct))/float64(inputPunct)

	return (paraScore + punctScore) / 2
}

// 辅助函数
func cleanWord(word string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return r
		}
		return -1
	}, word)
}

func isValidTerm(term string) bool {
	length := len(term)
	return length >= minWordLength && length <= maxWordLength
}

func isQualityResponse(content string) bool {
	// 基于长度、结构和关键词判断响应质量
	words := strings.Fields(content)
	if len(words) < 20 {
		return false
	}

	// 检查是否包含专业术语
	terms := extractProfessionalTerms(content)
	if len(terms) < 3 {
		return false
	}

	// 检查段落结构
	paragraphs := strings.Split(content, "\n\n")
	return len(paragraphs) >= 2
}

func extractProfessionalTerms(text string) []string {
	// 这里可以维护一个专业术语词典
	// 简化版本：检查长单词和特定前缀
	var terms []string
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		word = cleanWord(word)
		if len(word) > 8 || hasTechnicalPrefix(word) {
			terms = append(terms, word)
		}
	}

	return terms
}

func hasTechnicalPrefix(word string) bool {
	prefixes := []string{"micro", "multi", "inter", "cyber", "tech", "auto", "meta"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(word, prefix) {
			return true
		}
	}
	return false
}

func countPunctuation(text string) int {
	count := 0
	for _, r := range text {
		if unicode.IsPunct(r) {
			count++
		}
	}
	return count
}

func normalizeScore(score float64) float64 {
	// 确保分数在0.1-1.0之间
	if score < 0.1 {
		return 0.1
	}
	if score > 1.0 {
		return 1.0
	}
	return score
}
