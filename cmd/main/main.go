package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Koreychenko/go-quest/quest"
	"github.com/Koreychenko/go-quest/quest/pkg/bot"
	"github.com/Koreychenko/go-quest/quest/pkg/gemini"
	"github.com/Koreychenko/go-quest/quest/pkg/telegram"
	_ "github.com/joho/godotenv/autoload"
)

const (
	main_bot  = "main"
	liza_bot  = "liza"
	katya_bot = "katya"
)

const (
	stateFinish = "finish"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx := context.Background()

	// Load configuration
	cfg := quest.LoadConfig("./config.yaml")

	aiClient, err := gemini.NewClient(
		cfg.LLMAPIKey,
		quest.WithModelName(cfg.GenerationConfig.ModelName),
		quest.WithTemperature(cfg.GenerationConfig.Temperature),
		quest.WithMaxOutputTokens(cfg.GenerationConfig.MaxOutputTokens),
	)

	if err != nil {
		logger.Error("failed to create GenerationConfig client", "error", err)

		os.Exit(1)
	}

	if err = aiClient.ValidateAPIKey(ctx); err != nil {
		logger.Error("generationConfig API key validation failed", "error", err)

		os.Exit(1)
	}

	botManager := bot.NewManager(logger, aiClient)
	attachHandlers(cfg)
	err = botManager.AddBots(cfg)

	if err != nil {
		logger.Error("failed to add bots", "error", err)

		os.Exit(1)
	}

	tgManager := telegram.NewManager(logger, botManager)
	err = tgManager.AddBots(cfg)

	if err != nil {
		logger.Error("failed to add bots", "error", err)

		os.Exit(1)
	}

	engine, err := quest.NewGameEngine(botManager, tgManager, logger)
	if err != nil {
		logger.Error("Failed to initialize service", "error", err)

		os.Exit(1)
	}

	if err = engine.Run(ctx); err != nil {
		logger.Error("GameEngine error", "error", err)
		os.Exit(1)
	}
}

func attachHandlers(cfg *quest.Config) {
	cfg.Bots[main_bot].AddCommandHandler("start", mainBotStartCommandHandler(cfg.Bots[main_bot].Placeholders["liza_bot_name"]))
	cfg.Bots[main_bot].AddCommandHandler("restart", mainBotRestartCommandHandler(cfg.Bots[main_bot].Placeholders["liza_bot_name"]))
	cfg.Bots[main_bot].AddTransitionHandler(stateFinish, func(chatID int64, sender quest.MessageSender) error {
		time.Sleep(10 * time.Second)
		_ = sender.SendMessage(chatID, getMainBotFinalText())
		_ = sender.SendMessage(chatID, getShareLink(cfg.Bots[main_bot].Placeholders["main_bot_name"]))

		return nil
	})

	cfg.Bots[liza_bot].AddTransitionHandler(stateFinish, func(chatID int64, sender quest.MessageSender) error {
		time.Sleep(10 * time.Second)
		return sender.SendMessage(chatID, getFinalDecisionText())
	})

	cfg.Bots[katya_bot].AddCommandHandler("sendStarPhoto", func(chatID int64, sender quest.MessageSender) error {
		return sender.SendPhoto(chatID, "photos/katya_zvezda_school.png")
	})

	cfg.Bots[katya_bot].AddCommandHandler("sendChristmasTreePhoto", func(chatID int64, sender quest.MessageSender) error {
		return sender.SendPhoto(chatID, "photos/katya_elka_noch.png")
	})

	cfg.Bots[katya_bot].AddTransitionHandler(stateFinish, func(chatID int64, sender quest.MessageSender) error {
		time.Sleep(20 * time.Second)

		return sender.SendMessage(chatID, getKatyaFinalReplica())
	})
}

func mainBotStartCommandHandler(lizaBotName string) func(chatID int64, sender quest.MessageSender) error {
	return func(chatID int64, sender quest.MessageSender) error {
		err := sender.SendPhoto(chatID, "./photos/main_bot_photo1.png")

		if err != nil {
			slog.Error("unable to send photo:", "error", err.Error())
		}

		err = sender.SendMessage(chatID, getIntroText(lizaBotName))

		if err != nil {
			return err
		}

		err = sender.SendMessage(chatID, getHelperText())

		if err != nil {
			slog.Error("unable to send helper text:", "error", err.Error())
		}

		return nil
	}
}

func mainBotRestartCommandHandler(lizaBotName string) func(chatID int64, sender quest.MessageSender) error {
	return func(chatID int64, sender quest.MessageSender) error {
		_ = sender.SendMessage(chatID, "Игра перезапущена")

		return mainBotStartCommandHandler(lizaBotName)(chatID, sender)
	}
}

func getIntroText(lizaBotName string) string {
	return fmt.Sprintf(`
<b>7 января 1852 года</b> в Санкт-Петербурге в помещении Екатерининского вокзала была наряжена <b>первая в России</b> общественная ёлка.

После этого традиция наряжать ёлки для всех желающих распространилась по всей стране.

После реконструкции в 2011 году большую новогоднюю ёлку стали устанавливать в <a href="https://maps.app.goo.gl/Py7phmbcyQGXZiKbA">Новой Голландии</a>.

По легенде в <a href="https://maps.app.goo.gl/N7fa6cPezFXHHTFy8">Гимназии Петербургской культуры № 32</a> хранится <b>та самая звезда</b>, которая была зажжена на той первой ёлке, и которая дала начало новогодней традиции во всей стране.

Каждый год с момента реконструкции Новой Голландии Гимназия участвует в торжественном зажжении новогодней ёлки и передает свою драгоценную звезду для украшения, как символ волшебства и преемственности традиций.

Ученица 8-го класса Гимназии Лиза Волкова в этом году назначена ответственной за организацию торжественного в этом году. Она хотела, чтобы все прошло идеально.

<b>Но звезда внезапно пропала!</b>

Сможешь ли ты понять где она, кто её взял и как её вернуть?

Напиши Лизе! Ей очень нужна твоя помощь. До праздника остаются считанные дни!

%s
`, lizaBotName)
}

func getHelperText() string {
	return `
<i>Если что-то будет не понятно и не получаться, ты можешь задать свой вопрос здесь и я попробую тебе помочь.
Но будет намного интереснее додуматься до решения самостоятельно.</i>
`
}

func getFinalDecisionText() string {
	return `Короче, мы поговорили с директрисой и решили, что мы не должны забирать Звезду из больницы перед новым годом.

Тем более, что у них там <b>все начали выздоравливать</b> после того, как она у них появилась.

<b>Там она явно нужнее.</b>

А для Новой Голландии мы найдем другую звезду.

Спасибо тебе за помощь. И с Новым Годом!
`
}

func getKatyaFinalReplica() string {
	return `Привет, ты не поверишь! Сейчас пришел мой доктор и сказал, что меня смогут выписать перед Новым Годом! Это настоящее новогоднее чудо! Может быть это звезда помогла!`
}

func getMainBotFinalText() string {
	return `
Вот как иногда бывает, если очень сильно верить в чудо, оно может произойти.

Сегодня ты не просто помог найти пропавшую звезду, но и стал свидетелем начала новой доброй традиции.

<b>С наступающим Новым Годом!</b>
`
}

func getShareLink(mainBotName string) string {
	return fmt.Sprintf(`
Понравился квест?

Поделись с другом ссылкой на этого бота, чтобы он тоже смог поиграть: %s
`, mainBotName)
}
