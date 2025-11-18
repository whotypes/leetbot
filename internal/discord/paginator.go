package discord

import (
	"fmt"
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
	embed := &discordgo.MessageEmbed{}
	state.paginator.PageFunc(state.currentPage, embed)

	components := m.createButtons(state.messageID, state.currentPage, state.paginator.MaxPages)

	embeds := []*discordgo.MessageEmbed{embed}
	_, err := s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    state.channelID,
		ID:         state.messageID,
		Embeds:     &embeds,
		Components: &components,
	})
	return err
}

func (m *Manager) CreateInteraction(s *discordgo.Session, i *discordgo.Interaction, pg *Paginator, ephemeral bool) error {
	embed := &discordgo.MessageEmbed{}
	pg.PageFunc(0, embed)

	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

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
		return err
	}

	msg, err := s.InteractionResponse(i)
	if err != nil {
		return err
	}

	components = m.createButtons(msg.ID, 0, pg.MaxPages)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    msg.ChannelID,
		ID:         msg.ID,
		Components: &components,
	})
	if err != nil {
		return err
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

	return nil
}

func (m *Manager) CreateMessage(s *discordgo.Session, channelID string, pg *Paginator) error {
	embed := &discordgo.MessageEmbed{}
	pg.PageFunc(0, embed)

	tempID := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	components := m.createButtons(tempID, 0, pg.MaxPages)

	msg, err := s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return err
	}

	components = m.createButtons(msg.ID, 0, pg.MaxPages)
	_, err = s.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Channel:    channelID,
		ID:         msg.ID,
		Components: &components,
	})
	if err != nil {
		return err
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

	return nil
}

func (m *Manager) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}

	customID := i.MessageComponentData().CustomID
	if len(customID) < 10 || customID[:10] != "paginator:" {
		return
	}

	m.mu.RLock()
	state, exists := m.paginators[i.Message.ID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	var userID string
	if i.Member != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}

	if state.userID != "" && state.userID != userID {
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: m.notYourPaginatorMessage,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		if err != nil {
			fmt.Printf("Error responding to interaction: %v\n", err)
		}
		return
	}

	messageID := i.Message.ID
	var newPage int
	switch {
	case customID == fmt.Sprintf("paginator:%s:first", messageID):
		newPage = 0
	case customID == fmt.Sprintf("paginator:%s:back", messageID):
		newPage = state.currentPage - 1
		if newPage < 0 {
			newPage = 0
		}
	case customID == fmt.Sprintf("paginator:%s:next", messageID):
		newPage = state.currentPage + 1
		if newPage >= state.paginator.MaxPages {
			newPage = state.paginator.MaxPages - 1
		}
	case customID == fmt.Sprintf("paginator:%s:last", messageID):
		newPage = state.paginator.MaxPages - 1
	default:
		return
	}

	state.currentPage = newPage

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		fmt.Printf("Error responding to interaction: %v\n", err)
		return
	}

	err = m.updateMessage(s, state)
	if err != nil {
		fmt.Printf("Error updating paginator message: %v\n", err)
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
