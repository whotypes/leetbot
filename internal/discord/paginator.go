package discord

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/whotypes/leetbot/internal/data"
)

const (
	problemsPerPage     = 10
	paginationThreshold = 10
)

type Paginator struct {
	PageFunc func(page int, embed *discordgo.MessageEmbed)
	MaxPages int
}

type paginatorState struct {
	paginator *Paginator
	userID    string
	messageID string
	channelID string
	currentPage int
}

type Manager struct {
	mu        sync.RWMutex
	paginators map[string]*paginatorState
	notYourPaginatorMessage string
}

var PaginatorManager *Manager

func init() {
	PaginatorManager = &Manager{
		paginators: make(map[string]*paginatorState),
		notYourPaginatorMessage: "This paginator can only be used by the person who requested it.",
	}
}

func shouldUsePagination(problemCount int) bool {
	return problemCount > paginationThreshold
}

func createProblemsPaginator(company, timeframe string, problems []data.Problem) *Paginator {
	totalPages := (len(problems) + problemsPerPage - 1) / problemsPerPage

	return &Paginator{
		PageFunc: func(page int, embed *discordgo.MessageEmbed) {
			if page < 0 {
				page = 0
			}
			if page >= totalPages && totalPages > 0 {
				page = totalPages - 1
			}

			start := page * problemsPerPage
			end := start + problemsPerPage
			if end > len(problems) {
				end = len(problems)
			}

			pageProblems := problems[start:end]

			embed.Title = fmt.Sprintf("Most Popular Problems for %s (%s)",
				formatCompanyName(company),
				formatTimeframeDisplay(timeframe))
			embed.Color = 0x5865F2
			embed.Footer = &discordgo.MessageEmbedFooter{
				Text: fmt.Sprintf("Page %d/%d • Total: %d problems", page+1, totalPages, len(problems)),
			}
			embed.Timestamp = time.Now().Format(time.RFC3339)

			description := ""
			for i, problem := range pageProblems {
				problemNumber := start + i + 1
				difficultyIndicator := getDifficultyIndicator(problem.Difficulty)
				description += fmt.Sprintf("**%d.** %s [%s](<%s>) `%.0f%%`\n",
					problemNumber,
					difficultyIndicator,
					problem.Title,
					problem.URL,
					problem.Frequency)
			}

			embed.Description = description
		},
		MaxPages: totalPages,
	}
}

func (m *Manager) createButtons(messageID string, page, maxPages int) []discordgo.MessageComponent {
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{Name: "⏮"},
					Style: discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("paginator:%s:first", messageID),
					Disabled: page == 0,
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{Name: "◀"},
					Style: discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("paginator:%s:back", messageID),
					Disabled: page == 0,
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{Name: "▶"},
					Style: discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("paginator:%s:next", messageID),
					Disabled: page >= maxPages-1,
				},
				discordgo.Button{
					Emoji: &discordgo.ComponentEmoji{Name: "⏩"},
					Style: discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("paginator:%s:last", messageID),
					Disabled: page >= maxPages-1,
				},
			},
		},
	}
	return components
}

func (m *Manager) updateMessage(s *discordgo.Session, state *paginatorState) error {
	log.Printf("[PAGINATOR] Updating message %s to page %d/%d", state.messageID, state.currentPage+1, state.paginator.MaxPages)

	embed := &discordgo.MessageEmbed{}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PAGINATOR] PANIC in PageFunc for message %s: %v", state.messageID, r)
		}
	}()

	state.paginator.PageFunc(state.currentPage, embed)

	components := m.createButtons(state.messageID, state.currentPage, state.paginator.MaxPages)

	embeds := []*discordgo.MessageEmbed{embed}
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    state.channelID,
		ID:         state.messageID,
		Embeds:     &embeds,
		Components: &components,
	})

	if err != nil {
		log.Printf("[PAGINATOR] ERROR updating message %s: %v (channel: %s, page: %d/%d)",
			state.messageID, err, state.channelID, state.currentPage+1, state.paginator.MaxPages)
		return fmt.Errorf("failed to update message %s: %w", state.messageID, err)
	}

	log.Printf("[PAGINATOR] Successfully updated message %s to page %d/%d", state.messageID, state.currentPage+1, state.paginator.MaxPages)
	return nil
}

func (m *Manager) CreateInteraction(s *discordgo.Session, i *discordgo.Interaction, pg *Paginator, ephemeral bool) error {
	log.Printf("[PAGINATOR] Creating interaction paginator (maxPages: %d, ephemeral: %v)", pg.MaxPages, ephemeral)

	embed := &discordgo.MessageEmbed{}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PAGINATOR] PANIC in PageFunc during CreateInteraction: %v", r)
		}
	}()

	pg.PageFunc(0, embed)

	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}
	log.Printf("[PAGINATOR] User ID: %s", userID)

	var flags discordgo.MessageFlags
	if ephemeral {
		flags = discordgo.MessageFlagsEphemeral
	}

	tempID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	components := m.createButtons(tempID, 0, pg.MaxPages)

	err := s.InteractionRespond(i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      flags,
		},
	})
	if err != nil {
		log.Printf("[PAGINATOR] ERROR responding to interaction: %v", err)
		return fmt.Errorf("failed to respond to interaction: %w", err)
	}

	msg, err := s.InteractionResponse(i)
	if err != nil {
		log.Printf("[PAGINATOR] ERROR getting interaction response: %v", err)
		return fmt.Errorf("failed to get interaction response: %w", err)
	}
	log.Printf("[PAGINATOR] Created message %s in channel %s", msg.ID, msg.ChannelID)

	components = m.createButtons(msg.ID, 0, pg.MaxPages)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    msg.ChannelID,
		ID:         msg.ID,
		Components: &components,
	})
	if err != nil {
		log.Printf("[PAGINATOR] ERROR updating components for message %s: %v", msg.ID, err)
		return fmt.Errorf("failed to update message components: %w", err)
	}

	m.mu.Lock()
	m.paginators[msg.ID] = &paginatorState{
		paginator:   pg,
		userID:      userID,
		messageID:   msg.ID,
		channelID:   msg.ChannelID,
		currentPage: 0,
	}
	m.mu.Unlock()

	log.Printf("[PAGINATOR] Registered paginator for message %s (user: %s, pages: %d)", msg.ID, userID, pg.MaxPages)

	return nil
}

func (m *Manager) CreateMessage(s *discordgo.Session, channelID string, pg *Paginator) error {
	log.Printf("[PAGINATOR] Creating message paginator in channel %s (maxPages: %d)", channelID, pg.MaxPages)

	embed := &discordgo.MessageEmbed{}

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PAGINATOR] PANIC in PageFunc during CreateMessage: %v", r)
		}
	}()

	pg.PageFunc(0, embed)

	tempID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	components := m.createButtons(tempID, 0, pg.MaxPages)

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		log.Printf("[PAGINATOR] ERROR sending message to channel %s: %v", channelID, err)
		return fmt.Errorf("failed to send message: %w", err)
	}
	log.Printf("[PAGINATOR] Created message %s in channel %s", msg.ID, channelID)

	components = m.createButtons(msg.ID, 0, pg.MaxPages)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelID,
		ID:         msg.ID,
		Components: &components,
	})
	if err != nil {
		log.Printf("[PAGINATOR] ERROR updating components for message %s: %v", msg.ID, err)
		return fmt.Errorf("failed to update message components: %w", err)
	}

	m.mu.Lock()
	m.paginators[msg.ID] = &paginatorState{
		paginator:   pg,
		userID:      "",
		messageID:   msg.ID,
		channelID:   channelID,
		currentPage: 0,
	}
	m.mu.Unlock()

	log.Printf("[PAGINATOR] Registered paginator for message %s (pages: %d)", msg.ID, pg.MaxPages)

	return nil
}

func (m *Manager) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	log.Printf("[PAGINATOR] Received interaction: customID=%s, messageID=%s", customID, i.Message.ID)

	if len(customID) < 10 || customID[:10] != "paginator:" {
		log.Printf("[PAGINATOR] Ignoring non-paginator interaction: %s", customID)
		return
	}

	m.mu.RLock()
	state, exists := m.paginators[i.Message.ID]
	m.mu.RUnlock()

	if !exists {
		log.Printf("[PAGINATOR] WARNING: No paginator state found for message %s (customID: %s). Available paginators: %d",
			i.Message.ID, customID, len(m.paginators))
		m.mu.RLock()
		for msgID := range m.paginators {
			log.Printf("[PAGINATOR]   - Registered message ID: %s", msgID)
		}
		m.mu.RUnlock()
		return
	}

	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}
	log.Printf("[PAGINATOR] Processing interaction: messageID=%s, userID=%s, stateUserID=%s, currentPage=%d/%d",
		i.Message.ID, userID, state.userID, state.currentPage+1, state.paginator.MaxPages)

	if state.userID != "" && state.userID != userID {
		log.Printf("[PAGINATOR] User mismatch: expected %s, got %s", state.userID, userID)
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: m.notYourPaginatorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			log.Printf("[PAGINATOR] ERROR responding to user mismatch: %v", err)
		}
		return
	}

	messageID := i.Message.ID
	var newPage int
	var action string
	var currentPage int

	m.mu.RLock()
	currentPage = state.currentPage
	maxPages := state.paginator.MaxPages
	m.mu.RUnlock()

	switch {
	case customID == fmt.Sprintf("paginator:%s:first", messageID):
		newPage = 0
		action = "first"
	case customID == fmt.Sprintf("paginator:%s:back", messageID):
		newPage = currentPage - 1
		if newPage < 0 {
			newPage = 0
		}
		action = "back"
	case customID == fmt.Sprintf("paginator:%s:next", messageID):
		newPage = currentPage + 1
		if newPage >= maxPages {
			newPage = maxPages - 1
		}
		action = "next"
	case customID == fmt.Sprintf("paginator:%s:last", messageID):
		newPage = maxPages - 1
		action = "last"
	default:
		log.Printf("[PAGINATOR] WARNING: Unknown customID format: %s (expected format: paginator:%s:{action})", customID, messageID)
		log.Printf("[PAGINATOR]   Expected IDs: first=%s, back=%s, next=%s, last=%s",
			fmt.Sprintf("paginator:%s:first", messageID),
			fmt.Sprintf("paginator:%s:back", messageID),
			fmt.Sprintf("paginator:%s:next", messageID),
			fmt.Sprintf("paginator:%s:last", messageID))
		return
	}

	log.Printf("[PAGINATOR] Action: %s, moving from page %d to %d", action, currentPage+1, newPage+1)

	m.mu.Lock()
	state.currentPage = newPage
	m.mu.Unlock()

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		log.Printf("[PAGINATOR] ERROR responding to interaction (DeferredMessageUpdate): %v (messageID: %s, customID: %s)",
			err, messageID, customID)
		log.Printf("[PAGINATOR]   This error will cause 'This interaction failed' message")
		log.Printf("[PAGINATOR]   Interaction details: Type=%d, Token=%s", i.Type, i.Token)
		return
	}
	log.Printf("[PAGINATOR] Successfully deferred message update for message %s", messageID)

	err = m.updateMessage(s, state)
	if err != nil {
		log.Printf("[PAGINATOR] ERROR updating message after deferred response: %v (messageID: %s)", err, messageID)
	}
}

func sendPaginatedProblems(s *discordgo.Session, i *discordgo.InteractionCreate, company, timeframe string, problems []data.Problem) error {
	pg := createProblemsPaginator(company, timeframe, problems)

	return PaginatorManager.CreateInteraction(s, i.Interaction, pg, false)
}

func sendPaginatedProblemsMessage(s *discordgo.Session, channelID, company, timeframe string, problems []data.Problem) error {
	pg := createProblemsPaginator(company, timeframe, problems)

	return PaginatorManager.CreateMessage(s, channelID, pg)
}
