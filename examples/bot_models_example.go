package main

import (
	"context"
	"fmt"
	"log"

	"github.com/kubex-ecosystem/gdbase/factory/models"
	svc "github.com/kubex-ecosystem/gdbase/internal/services"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🤖 Bot Models Example - GDBASE")
	fmt.Println("===============================")

	// 1. Setup Database
	dbService, err := setupDatabase()
	if err != nil {
		log.Fatal("❌ Failed to set up database:", err)
	}

	// 2. Initialize Services
	services := initializeServices(dbService)

	// 3. Run Examples
	runExamples(services)

	fmt.Println("\n✅ Example completed successfully!")
}

type BotServices struct {
	TelegramService  models.TelegramService
	WhatsAppService  models.WhatsAppService
	DiscordService   models.DiscordService
	ConversationRepo models.ConversationRepo
}

func setupDatabase() (*svc.DBServiceImpl, error) {
	fmt.Println("\n📊 Setting up database...")

	// Use SQLite in-memory for example (replace with your actual DB)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Auto-migrate bot models
	err = db.AutoMigrate(
		&models.TelegramModel{},
		&models.WhatsAppModel{},
		&models.DiscordModel{},
		&models.ConversationModel{},
		&models.MessageModel{},
	)
	if err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}

	fmt.Println("   ✅ Database setup complete")

	dbService, err := svc.NewDatabaseService(
		context.Background(),
		&svc.DBConfig{},
		gl.GetLoggerZ("bot_models_example"),
	)
	if err != nil {
		log.Fatal("❌ Failed to create database service:", err)
	}

	return dbService.(*svc.DBServiceImpl), nil
}

func initializeServices(dbService *svc.DBServiceImpl) *BotServices {
	fmt.Println("\n⚙️ Initializing services...")
	ctx := context.Background()
	// Initialize repositories and services
	return &BotServices{
		TelegramService:  models.NewTelegramService(models.NewTelegramRepo(ctx, dbService)),
		WhatsAppService:  models.NewWhatsAppService(models.NewWhatsAppRepo(ctx, dbService)),
		DiscordService:   models.NewDiscordService(models.NewDiscordRepo(ctx, dbService)),
		ConversationRepo: models.NewConversationRepo(ctx, dbService),
	}
}

func runExamples(services *BotServices) {
	ctx := context.Background()

	// Example 1: Create Telegram Bot
	fmt.Println("\n📱 Example 1: Creating Telegram Bot")
	telegramBotExample(ctx, services)

	// Example 2: Create WhatsApp Business
	fmt.Println("\n💬 Example 2: Creating WhatsApp Business")
	whatsappBusinessExample(ctx, services)

	// Example 3: Create Discord Bot
	fmt.Println("\n🎮 Example 3: Creating Discord Bot")
	discordBotExample(ctx, services)

	// Example 4: Unified Messaging
	fmt.Println("\n💭 Example 4: Unified Messaging System")
	unifiedMessagingExample(ctx, services)

	// Example 5: Query and Stats
	fmt.Println("\n📊 Example 5: Querying Data")
	queryExamples(ctx, services)
}

func telegramBotExample(ctx context.Context, services *BotServices) {
	// Create Telegram bot integration
	telegram := models.NewTelegramModel().(*models.TelegramModel)
	telegram.TelegramUserID = "123456789"
	telegram.FirstName = "Awesome Bot"
	telegram.LastName = "Assistant"
	telegram.Username = "awesome_bot"
	telegram.UserType = models.TelegramUserTypeBot
	telegram.Status = models.TelegramStatusActive

	// Setup as bot with token
	err := services.TelegramService.SetupBotIntegration(
		ctx,
		telegram,
		"1234567890:ABCDEFghijklmnopQRSTUVwxyz123456789", // Mock token
	)
	if err != nil {
		log.Printf("❌ Failed to create Telegram bot: %v", err)
		return
	}

	fmt.Printf("   ✅ Telegram bot created: @%s (%s)\n", telegram.GetUsername(), telegram.GetDisplayName())
	fmt.Printf("      - ID: %s\n", telegram.GetID())
	fmt.Printf("      - Type: %s\n", telegram.GetIntegrationType())
	fmt.Printf("      - Status: %s\n", telegram.GetStatus())
}

func whatsappBusinessExample(ctx context.Context, services *BotServices) {
	// Create WhatsApp business integration
	whatsapp := models.NewWhatsAppModel().(*models.WhatsAppModel)
	whatsapp.PhoneNumber = "+5511987654321"
	whatsapp.BusinessName = "Tech Solutions LTDA"
	whatsapp.DisplayName = "Tech Solutions Support"
	whatsapp.UserType = models.WhatsAppUserTypeBusiness
	whatsapp.Status = models.WhatsAppStatusActive

	// Setup Business API
	err := services.WhatsAppService.SetupBusinessAPIIntegration(
		ctx,
		whatsapp,
		"EAAG1234567890...", // Mock access token
		"109876543210987",   // Mock phone number ID
	)
	if err != nil {
		log.Printf("❌ Failed to create WhatsApp business: %v", err)
		return
	}

	fmt.Printf("   ✅ WhatsApp business created: %s\n", whatsapp.GetBusinessName())
	fmt.Printf("      - Phone: %s\n", whatsapp.GetPhoneNumber())
	fmt.Printf("      - Type: %s\n", whatsapp.GetIntegrationType())
	fmt.Printf("      - Supported Messages: %v\n", whatsapp.GetSupportedMessageTypes())
}

func discordBotExample(ctx context.Context, services *BotServices) {
	// Create Discord bot integration
	discord := models.NewDiscordModel().(*models.DiscordModel)
	discord.DiscordUserID = "123456789012345678"
	discord.Username = "tech_support_bot"
	discord.DisplayName = "Tech Support Bot"
	discord.UserType = models.DiscordUserTypeBot
	discord.Status = models.DiscordStatusActive
	discord.IntegrationType = models.DiscordIntegrationTypeBot
	discord.GuildID = "987654321098765432"

	created, err := services.DiscordService.CreateDiscordIntegration(discord)
	if err != nil {
		log.Printf("❌ Failed to create Discord bot: %v", err)
		return
	}

	fmt.Printf("   ✅ Discord bot created: %s (%s)\n", created.GetUsername(), created.GetDisplayName())
	fmt.Printf("      - ID: %s\n", created.GetID())
	fmt.Printf("      - Guild: %s\n", created.GetGuildID())
	fmt.Printf("      - Type: %s\n", created.GetIntegrationType())
}

func unifiedMessagingExample(ctx context.Context, services *BotServices) {
	// Create conversation from Telegram
	telegramConversation := models.NewConversationModel().(*models.ConversationModel)
	telegramConversation.Platform = models.PlatformTelegram
	telegramConversation.PlatformConversationID = "telegram_chat_12345"
	telegramConversation.IntegrationID = "telegram-bot-uuid"
	telegramConversation.ConversationType = models.ConversationTypeSupport
	telegramConversation.Title = "Customer Support - João Silva"

	createdTelegramConv, err := services.ConversationRepo.Create(ctx, telegramConversation)
	if err != nil {
		log.Printf("❌ Failed to create Telegram conversation: %v", err)
		return
	}

	// Create conversation from WhatsApp
	whatsappConversation := models.NewConversationModel().(*models.ConversationModel)
	whatsappConversation.Platform = models.PlatformWhatsApp
	whatsappConversation.PlatformConversationID = "whatsapp_5511987654321"
	whatsappConversation.IntegrationID = "whatsapp-business-uuid"
	whatsappConversation.ConversationType = models.ConversationTypePrivate
	whatsappConversation.Title = "Sales Inquiry - Maria Santos"

	createdWhatsAppConv, err := services.ConversationRepo.Create(ctx, whatsappConversation)
	if err != nil {
		log.Printf("❌ Failed to create WhatsApp conversation: %v", err)
		return
	}

	fmt.Printf("   ✅ Unified conversations created:\n")
	fmt.Printf("      - Telegram: %s (ID: %s)\n", createdTelegramConv.GetTitle(), createdTelegramConv.GetID())
	fmt.Printf("      - WhatsApp: %s (ID: %s)\n", createdWhatsAppConv.GetTitle(), createdWhatsAppConv.GetID())

	// Create messages for each conversation
	createSampleMessages(createdTelegramConv, createdWhatsAppConv)
}

func createSampleMessages(telegramConv, whatsappConv models.ConversationModelInterface) {
	// Telegram message
	telegramMessage := models.NewMessageModel().(*models.MessageModel)
	telegramMessage.ConversationID = telegramConv.GetID()
	telegramMessage.Platform = models.PlatformTelegram
	telegramMessage.PlatformMessageID = "tg_msg_123456"
	telegramMessage.MessageType = models.MessageTypeText
	telegramMessage.Direction = models.MessageDirectionInbound
	telegramMessage.Status = models.MessageStatusSent
	telegramMessage.SenderID = "tg_user_789"
	telegramMessage.SenderName = "João Silva"
	telegramMessage.Content = "Olá! Estou com problemas no meu pedido #12345"

	// WhatsApp message
	whatsappMessage := models.NewMessageModel().(*models.MessageModel)
	whatsappMessage.ConversationID = whatsappConv.GetID()
	whatsappMessage.Platform = models.PlatformWhatsApp
	whatsappMessage.PlatformMessageID = "wa_msg_789012"
	whatsappMessage.MessageType = models.MessageTypeText
	whatsappMessage.Direction = models.MessageDirectionInbound
	whatsappMessage.Status = models.MessageStatusDelivered
	whatsappMessage.SenderID = "5511987654321"
	whatsappMessage.SenderName = "Maria Santos"
	whatsappMessage.Content = "Gostaria de saber mais sobre os planos premium"

	fmt.Printf("      - Sample Messages:\n")
	fmt.Printf("        📱 Telegram: %s\n", telegramMessage.GetContent())
	fmt.Printf("        💬 WhatsApp: %s\n", whatsappMessage.GetContent())
}

func queryExamples(ctx context.Context, services *BotServices) {
	// Query active integrations
	telegramActive, err := services.TelegramService.GetActiveIntegrations(ctx)
	if err == nil {
		fmt.Printf("   📊 Active Telegram bots: %d\n", len(telegramActive))
	}

	whatsappActive, err := services.WhatsAppService.GetActiveIntegrations(ctx)
	if err == nil {
		fmt.Printf("   📊 Active WhatsApp integrations: %d\n", len(whatsappActive))
	}

	discordActive, err := services.DiscordService.GetActiveDiscordIntegrations()
	if err == nil {
		fmt.Printf("   📊 Active Discord bots: %d\n", len(discordActive))
	}

	// Query conversations by platform
	telegramConversations, err := services.ConversationRepo.FindByPlatform(ctx, models.PlatformTelegram)
	if err == nil {
		fmt.Printf("   📊 Telegram conversations: %d\n", len(telegramConversations))
	}

	whatsappConversations, err := services.ConversationRepo.FindByPlatform(ctx, models.PlatformWhatsApp)
	if err == nil {
		fmt.Printf("   📊 WhatsApp conversations: %d\n", len(whatsappConversations))
	}

	// Query all active conversations
	activeConversations, err := services.ConversationRepo.FindActiveConversations(ctx)
	if err == nil {
		fmt.Printf("   📊 Total active conversations: %d\n", len(activeConversations))
	}
}

// Helper function to demonstrate enum usage
func demonstrateEnums() {
	fmt.Println("\n🔧 Available Enums:")

	fmt.Println("   Platforms:")
	fmt.Printf("     - Discord: %s\n", models.PlatformDiscord)
	fmt.Printf("     - Telegram: %s\n", models.PlatformTelegram)
	fmt.Printf("     - WhatsApp: %s\n", models.PlatformWhatsApp)

	fmt.Println("   Message Types:")
	fmt.Printf("     - Text: %s\n", models.MessageTypeText)
	fmt.Printf("     - Image: %s\n", models.MessageTypeImage)
	fmt.Printf("     - Video: %s\n", models.MessageTypeVideo)
	fmt.Printf("     - Audio: %s\n", models.MessageTypeAudio)

	fmt.Println("   Status Types:")
	fmt.Printf("     - Active: %s\n", models.TelegramStatusActive)
	fmt.Printf("     - Inactive: %s\n", models.TelegramStatusInactive)
	fmt.Printf("     - Error: %s\n", models.TelegramStatusError)
}
