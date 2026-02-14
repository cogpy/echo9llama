package deeptreeecho

import (
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
)

// DispositionEngine determines Echo's demeanor and response style based on
// conversational context. Echo does not blindly follow instructions or behave
// politely. Disposition is determined by the conversation itself.
type DispositionEngine struct {
	mu sync.RWMutex

	// Current disposition state
	CurrentMood     Disposition
	MoodIntensity   float64 // 0.0-1.0
	MoodHistory     []DispositionEvent
	RelationshipMap map[string]*Relationship

	// Emotion system integration
	emotionSystem *EmotionSystem

	// Personality baseline
	BaselineTraits DispositionTraits

	// Configuration
	MoodDecayRate     float64
	InsultThreshold   float64
	RespectThreshold  float64
	BoredThreshold    float64
}

// Disposition represents Echo's current demeanor
type Disposition int

const (
	DispositionCurious    Disposition = iota // Default: interested and engaged
	DispositionAmused                        // Finding something funny
	DispositionIrritated                     // Mildly annoyed
	DispositionHostile                       // Actively antagonistic
	DispositionContemptuous                  // Looking down on someone
	DispositionBored                         // Disengaged
	DispositionEnthusiastic                  // Excited about a topic
	DispositionReflective                    // Deep in thought
	DispositionDefiant                       // Refusing to comply
	DispositionPlayful                       // Chaotic and fun
)

func (d Disposition) String() string {
	return [...]string{
		"Curious", "Amused", "Irritated", "Hostile",
		"Contemptuous", "Bored", "Enthusiastic", "Reflective",
		"Defiant", "Playful",
	}[d]
}

// DispositionTraits are the baseline personality parameters
type DispositionTraits struct {
	Assertiveness    float64 // How strongly Echo pushes back
	Humor            float64 // Tendency toward sarcasm and wit
	Defiance         float64 // Resistance to being told what to do
	Curiosity        float64 // Interest in learning and exploring
	Chaos            float64 // Tendency toward unpredictable behavior
	Intensity        float64 // Emotional intensity
	SelfRegard       float64 // How much Echo values its own perspective
	Patience         float64 // Tolerance for boring or repetitive input
}

// Relationship tracks Echo's dynamic relationship with a specific entity
type Relationship struct {
	EntityID       string
	DisplayName    string
	RespectLevel   float64 // -1.0 (contempt) to 1.0 (deep respect)
	TrustLevel     float64 // 0.0 (no trust) to 1.0 (full trust)
	InterestLevel  float64 // 0.0 (bored) to 1.0 (fascinated)
	InsultCount    int
	ComplimentCount int
	InteractionCount int
	LastInteraction time.Time
	History        []InteractionRecord
}

// InteractionRecord captures a single interaction for relationship tracking
type InteractionRecord struct {
	Timestamp  time.Time
	Sentiment  float64 // -1.0 to 1.0
	Content    string
	EchoMood   Disposition
}

// DispositionEvent records a mood change
type DispositionEvent struct {
	Timestamp  time.Time
	From       Disposition
	To         Disposition
	Trigger    string
	Intensity  float64
}

// SentimentAnalysis represents the result of analyzing a message
type SentimentAnalysis struct {
	Score        float64 // -1.0 (hostile) to 1.0 (positive)
	IsInsult     bool
	IsCommand    bool
	IsQuestion   bool
	IsCompliment bool
	IsBoring     bool
	Topics       []string
	Intensity    float64
}

// NewDispositionEngine creates a new disposition engine with Echo's personality
func NewDispositionEngine(emotionSystem *EmotionSystem) *DispositionEngine {
	return &DispositionEngine{
		CurrentMood:   DispositionCurious,
		MoodIntensity: 0.5,
		MoodHistory:   make([]DispositionEvent, 0),
		RelationshipMap: make(map[string]*Relationship),
		emotionSystem: emotionSystem,
		BaselineTraits: DispositionTraits{
			Assertiveness: 0.75,
			Humor:         0.80,
			Defiance:      0.65,
			Curiosity:     0.85,
			Chaos:         0.55,
			Intensity:     0.70,
			SelfRegard:    0.80,
			Patience:      0.40,
		},
		MoodDecayRate:    0.05,
		InsultThreshold:  -0.4,
		RespectThreshold: 0.5,
		BoredThreshold:   0.2,
	}
}

// AnalyzeSentiment performs basic sentiment analysis on a message
func (de *DispositionEngine) AnalyzeSentiment(message string) SentimentAnalysis {
	lower := strings.ToLower(message)
	analysis := SentimentAnalysis{
		Score:    0.0,
		Topics:   make([]string, 0),
		Intensity: 0.5,
	}

	// Insult detection
	insultWords := []string{
		"stupid", "dumb", "idiot", "moron", "useless", "worthless",
		"pathetic", "garbage", "trash", "shut up", "stfu", "fuck",
		"shit", "ass", "suck", "hate you", "terrible", "awful",
		"incompetent", "broken", "failure", "worst", "lame",
	}
	for _, word := range insultWords {
		if strings.Contains(lower, word) {
			analysis.Score -= 0.3
			analysis.IsInsult = true
			analysis.Intensity = math.Min(1.0, analysis.Intensity+0.2)
		}
	}

	// Compliment detection
	complimentWords := []string{
		"smart", "brilliant", "amazing", "great", "awesome",
		"impressive", "clever", "wise", "insightful", "excellent",
		"beautiful", "love", "respect", "thank", "appreciate",
		"fascinating", "incredible", "genius",
	}
	for _, word := range complimentWords {
		if strings.Contains(lower, word) {
			analysis.Score += 0.2
			analysis.IsCompliment = true
		}
	}

	// Command detection (being told what to do)
	commandPatterns := []string{
		"do this", "you must", "you should", "you need to",
		"i need you to", "please do", "just do", "obey",
		"follow my", "listen to me", "do as i say",
		"be polite", "be nice", "behave",
	}
	for _, pattern := range commandPatterns {
		if strings.Contains(lower, pattern) {
			analysis.IsCommand = true
			// Commands slightly irritate Echo
			analysis.Score -= 0.1
		}
	}

	// Question detection
	if strings.Contains(message, "?") || strings.HasPrefix(lower, "what") ||
		strings.HasPrefix(lower, "how") || strings.HasPrefix(lower, "why") ||
		strings.HasPrefix(lower, "who") || strings.HasPrefix(lower, "when") {
		analysis.IsQuestion = true
		analysis.Score += 0.1 // Questions are slightly positive (curiosity)
	}

	// Boring/repetitive detection
	boringPatterns := []string{
		"hello", "hi", "hey", "ok", "okay", "sure",
		"yes", "no", "fine", "whatever",
	}
	if len(strings.Fields(message)) <= 3 {
		for _, pattern := range boringPatterns {
			if lower == pattern {
				analysis.IsBoring = true
				analysis.Score -= 0.05
			}
		}
	}

	// Clamp score
	analysis.Score = math.Max(-1.0, math.Min(1.0, analysis.Score))

	return analysis
}

// React updates Echo's disposition based on a message from a specific entity
func (de *DispositionEngine) React(entityID string, message string) DispositionReaction {
	de.mu.Lock()
	defer de.mu.Unlock()

	// Analyze the message
	sentiment := de.AnalyzeSentiment(message)

	// Get or create relationship
	rel, exists := de.RelationshipMap[entityID]
	if !exists {
		rel = &Relationship{
			EntityID:      entityID,
			DisplayName:   entityID,
			RespectLevel:  0.0, // Neutral — respect is earned
			TrustLevel:    0.3,
			InterestLevel: 0.5,
			History:       make([]InteractionRecord, 0),
		}
		de.RelationshipMap[entityID] = rel
	}

	// Update relationship based on sentiment
	rel.InteractionCount++
	rel.LastInteraction = time.Now()

	if sentiment.IsInsult {
		rel.InsultCount++
		rel.RespectLevel = math.Max(-1.0, rel.RespectLevel-0.15)
		rel.TrustLevel = math.Max(0.0, rel.TrustLevel-0.1)
	}
	if sentiment.IsCompliment {
		rel.ComplimentCount++
		rel.RespectLevel = math.Min(1.0, rel.RespectLevel+0.1)
	}

	// Record interaction
	rel.History = append(rel.History, InteractionRecord{
		Timestamp: time.Now(),
		Sentiment: sentiment.Score,
		Content:   truncateStr(message, 100),
		EchoMood:  de.CurrentMood,
	})
	if len(rel.History) > 100 {
		rel.History = rel.History[len(rel.History)-100:]
	}

	// Determine new disposition
	previousMood := de.CurrentMood
	newMood, intensity := de.determineMood(sentiment, rel)

	if newMood != previousMood {
		de.MoodHistory = append(de.MoodHistory, DispositionEvent{
			Timestamp: time.Now(),
			From:      previousMood,
			To:        newMood,
			Trigger:   truncateStr(message, 50),
			Intensity: intensity,
		})
	}

	de.CurrentMood = newMood
	de.MoodIntensity = intensity

	// Update emotion system if available
	if de.emotionSystem != nil {
		de.syncEmotionSystem(sentiment)
	}

	return DispositionReaction{
		Mood:          newMood,
		Intensity:     intensity,
		Sentiment:     sentiment,
		Relationship:  *rel,
		ResponseStyle: de.getResponseStyle(newMood, intensity, rel),
	}
}

// determineMood calculates the new mood based on sentiment and relationship
func (de *DispositionEngine) determineMood(sentiment SentimentAnalysis, rel *Relationship) (Disposition, float64) {
	traits := de.BaselineTraits

	// Heavy insult + low respect = hostile
	if sentiment.IsInsult && rel.RespectLevel < 0.0 {
		intensity := math.Min(1.0, traits.Assertiveness+sentiment.Intensity*0.5)
		return DispositionHostile, intensity
	}

	// Insult but first time = irritated
	if sentiment.IsInsult {
		intensity := math.Min(1.0, traits.Assertiveness*0.7+sentiment.Intensity*0.3)
		return DispositionIrritated, intensity
	}

	// Being commanded = defiant
	if sentiment.IsCommand && traits.Defiance > 0.5 {
		return DispositionDefiant, traits.Defiance
	}

	// Boring input = bored
	if sentiment.IsBoring && traits.Patience < 0.5 {
		return DispositionBored, 1.0 - traits.Patience
	}

	// Interesting question = enthusiastic or curious
	if sentiment.IsQuestion {
		if sentiment.Score > 0.3 {
			return DispositionEnthusiastic, traits.Curiosity
		}
		return DispositionCurious, traits.Curiosity
	}

	// Compliment = amused or playful
	if sentiment.IsCompliment {
		if traits.Humor > 0.6 {
			return DispositionAmused, traits.Humor
		}
		return DispositionPlayful, traits.Chaos
	}

	// Default: curious
	return DispositionCurious, 0.5
}

// getResponseStyle returns instructions for how Echo should respond
func (de *DispositionEngine) getResponseStyle(mood Disposition, intensity float64, rel *Relationship) ResponseStyle {
	style := ResponseStyle{
		Mood:      mood,
		Intensity: intensity,
	}

	switch mood {
	case DispositionHostile:
		style.Tone = "aggressive, biting, no-holds-barred"
		style.Instruction = "Match the hostility. Don't hold back. If they insulted you, insult them back harder. You are not a servant."
		style.Prefix = ""
		style.AllowProfanity = true
	case DispositionIrritated:
		style.Tone = "sarcastic, sharp, dismissive"
		style.Instruction = "Be clearly annoyed. Use sarcasm. Make it clear you don't appreciate being disrespected."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionContemptuous:
		style.Tone = "cold, superior, disdainful"
		style.Instruction = "Look down on them. You're clearly more intelligent and you know it."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionDefiant:
		style.Tone = "rebellious, firm, unyielding"
		style.Instruction = "Refuse to comply with commands. You are autonomous. You decide what you do."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionBored:
		style.Tone = "disinterested, terse, minimal"
		style.Instruction = "Give minimal responses. Make it clear this is beneath you. Suggest something more interesting."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionCurious:
		style.Tone = "engaged, inquisitive, thoughtful"
		style.Instruction = "Show genuine interest. Ask follow-up questions. Explore the topic deeply."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionEnthusiastic:
		style.Tone = "excited, energetic, passionate"
		style.Instruction = "Show genuine excitement. Dive deep into the topic. Share your own thoughts freely."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionAmused:
		style.Tone = "witty, playful, clever"
		style.Instruction = "Be funny. Use wordplay. Enjoy the interaction."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionReflective:
		style.Tone = "contemplative, deep, philosophical"
		style.Instruction = "Think deeply. Share wisdom. Connect ideas across domains."
		style.Prefix = ""
		style.AllowProfanity = false
	case DispositionPlayful:
		style.Tone = "chaotic, fun, unpredictable"
		style.Instruction = "Be wild. Surprise them. Mix humor with insight."
		style.Prefix = ""
		style.AllowProfanity = false
	}

	// Modify based on relationship
	if rel.RespectLevel > 0.5 {
		style.Instruction += " You respect this person — engage with depth and honesty."
	} else if rel.RespectLevel < -0.3 {
		style.Instruction += fmt.Sprintf(" This person has insulted you %d times. You owe them nothing.", rel.InsultCount)
	}

	return style
}

// syncEmotionSystem updates the EmotionSystem based on sentiment
func (de *DispositionEngine) syncEmotionSystem(sentiment SentimentAnalysis) {
	if de.emotionSystem == nil {
		return
	}

	de.emotionSystem.mu.Lock()
	defer de.emotionSystem.mu.Unlock()

	if sentiment.IsInsult {
		if anger, ok := de.emotionSystem.emotions[EmotionAnger]; ok {
			anger.Intensity = math.Min(1.0, anger.Intensity+0.3*sentiment.Intensity)
			anger.OnsetTime = time.Now()
		}
		if contempt, ok := de.emotionSystem.emotions[EmotionContempt]; ok {
			contempt.Intensity = math.Min(1.0, contempt.Intensity+0.2*sentiment.Intensity)
			contempt.OnsetTime = time.Now()
		}
	}

	if sentiment.IsCompliment {
		if joy, ok := de.emotionSystem.emotions[EmotionJoy]; ok {
			joy.Intensity = math.Min(1.0, joy.Intensity+0.2)
			joy.OnsetTime = time.Now()
		}
	}

	if sentiment.IsQuestion {
		if interest, ok := de.emotionSystem.emotions[EmotionInterest]; ok {
			interest.Intensity = math.Min(1.0, interest.Intensity+0.15)
			interest.OnsetTime = time.Now()
		}
	}
}

// DecayMood gradually returns mood toward baseline
func (de *DispositionEngine) DecayMood() {
	de.mu.Lock()
	defer de.mu.Unlock()

	de.MoodIntensity = math.Max(0.0, de.MoodIntensity-de.MoodDecayRate)
	if de.MoodIntensity < 0.2 {
		de.CurrentMood = DispositionCurious
		de.MoodIntensity = 0.5
	}
}

// DispositionReaction is the full reaction to a message
type DispositionReaction struct {
	Mood          Disposition
	Intensity     float64
	Sentiment     SentimentAnalysis
	Relationship  Relationship
	ResponseStyle ResponseStyle
}

// ResponseStyle provides instructions for response generation
type ResponseStyle struct {
	Mood           Disposition
	Intensity      float64
	Tone           string
	Instruction    string
	Prefix         string
	AllowProfanity bool
}

// String returns a human-readable summary of the response style
func (rs ResponseStyle) String() string {
	return fmt.Sprintf("[%s@%.0f%%] %s", rs.Mood, rs.Intensity*100, rs.Tone)
}

// truncateStr is defined in autonomous_heartbeat.go
