package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/djalben/istok-agent-core/internal/application/usecases"
	"github.com/djalben/istok-agent-core/internal/domain"
	"gitlab.com/libs-artifex/wrapper/v2"
)

type agentModeRun struct {
	o                  *Orchestrator
	ctx                context.Context
	specification      string
	url                string
	startTime          time.Time
	result             *GenerationResult
	fsm                *domain.TaskStateMachine
	competitorFeatures []CompetitorFeature
	sessionID          string
	manifest           *SystemManifest
	masterPlan         *MasterPlan
	imageURLs          map[string]string
	generatedCode      map[string]string
	generateVideo      bool
}

func (o *Orchestrator) generateAgentMode(ctx context.Context, specification string, url string) (*GenerationResult, error) {
	run := &agentModeRun{
		o:             o,
		ctx:           ctx,
		specification: specification,
		url:           url,
		startTime:     time.Now(),
		result: &GenerationResult{
			Code:   make(map[string]string),
			Assets: make(map[string]string),
		},
		imageURLs:     map[string]string{},
		generateVideo: generateVideoFromContext(ctx),
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()
	run.ctx = ctx
	run.fsm = domain.NewTaskStateMachine()
	run.sessionID, _ = ctx.Value(sessionIDKey{}).(string)

	return run.execute()
}

func (run *agentModeRun) execute() (*GenerationResult, error) {
	err := run.phaseAgentResearch()
	if err != nil {
		return nil, err
	}
	err = run.phaseAgentArchitectureAndPlan()
	if err != nil {
		return nil, err
	}
	err = run.phaseAgentFeatureApproval()
	if err != nil {
		return nil, err
	}
	err = run.phaseAgentPostPlanFSM()
	if err != nil {
		return nil, err
	}
	run.phaseAgentDesigner()
	res, err := run.phaseAgentCoding()
	if err != nil && !errors.Is(err, ErrCodingPhaseDone) {
		return res, err
	}

	return run.phaseAgentVerification()
}

func (run *agentModeRun) phaseAgentResearch() error {
	applog(run.ctx).DebugContext(run.ctx, "fsm transition", "from", "created", "to", "researching")
	err := run.fsm.TransitionTo(domain.StateResearching, "starting research phase")
	if err != nil {
		applog(run.ctx).ErrorContext(run.ctx, "fsm transition failed", "from", "created", "to", "researching", "error", wrapper.Wrap(err))

		return wrapper.Wrap(err)
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateCreated, domain.StateResearching, "agent mode")

	researcher := NewResearcherAgent(run.o.llm)
	if run.url != "" {
		synthesis, _ := run.o.deepSynthesis(run.ctx, run.url, run.specification)
		if synthesis != nil && len(synthesis.Features) > 0 {
			run.competitorFeatures = synthesis.Features
		}

		visualAudit, err := researcher.VisualAudit(run.ctx, run.url, run.o.busFromCtx(run.ctx))
		if err != nil {
			run.o.sendStatus(run.ctx, RoleResearcher, "error", fmt.Sprintf("⚠️ URL-аудит недоступен: %v", err), 0)
		} else {
			run.o.mu.Lock()
			run.result.VisualAudit = visualAudit
			run.result.Audit = &ReverseEngineeringResult{
				URL:          run.url,
				Colors:       visualAudit.Colors,
				Fonts:        visualAudit.Fonts,
				Components:   visualAudit.Components,
				Layout:       visualAudit.Layout,
				Technologies: visualAudit.Technologies,
				Audit:        fmt.Sprintf("DesignSystem: %s, Animations: %v", visualAudit.DesignSystem, visualAudit.Animations),
			}
			run.o.mu.Unlock()
		}
	} else {
		visualAudit := researcher.AnalyzeSpec(run.ctx, run.specification, run.o.busFromCtx(run.ctx))
		run.o.mu.Lock()
		run.result.VisualAudit = visualAudit
		run.result.Audit = &ReverseEngineeringResult{
			URL:          "spec://text",
			Colors:       visualAudit.Colors,
			Fonts:        visualAudit.Fonts,
			Components:   visualAudit.Components,
			Layout:       visualAudit.Layout,
			Technologies: visualAudit.Technologies,
			Audit:        fmt.Sprintf("DesignSystem: %s, Insights: %v", visualAudit.DesignSystem, visualAudit.Insights),
		}
		run.o.mu.Unlock()
	}

	applog(run.ctx).DebugContext(run.ctx, "fsm transition", "from", "researching", "to", "planning")
	err = run.fsm.TransitionTo(domain.StatePlanning, "research complete, starting planning")
	if err != nil {
		applog(run.ctx).ErrorContext(run.ctx, "fsm transition failed", "from", "researching", "to", "planning", "error", wrapper.Wrap(err))

		return wrapper.Wrap(err)
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateResearching, domain.StatePlanning, "research done")

	return nil
}

func (run *agentModeRun) phaseAgentArchitectureAndPlan() error {
	applog(run.ctx).DebugContext(run.ctx, "architect phase start")
	var archErr error
	run.manifest, archErr = run.o.defineArchitecture(run.ctx, run.specification, run.result.Audit, run.competitorFeatures)
	if archErr != nil {
		applog(run.ctx).WarnContext(run.ctx, "architecture manifest warning", "error", wrapper.Wrap(archErr))
	} else {
		applog(run.ctx).DebugContext(run.ctx, "architect success", "hasManifest", run.manifest != nil)
	}

	applog(run.ctx).DebugContext(run.ctx, "brain phase start")
	run.o.sendStatus(run.ctx, RoleBrain, "running", "🧠 Стратег Истока анализирует архитектуру...", 18)
	strategy, brainErr := run.o.synthesizeStrategy(run.ctx, run.specification, run.result.Audit)
	if brainErr != nil {
		applog(run.ctx).WarnContext(run.ctx, "brain synthesis warning", "error", wrapper.Wrap(brainErr))
	} else {
		applog(run.ctx).DebugContext(run.ctx, "brain success", "strategyLen", len(strategy))
		if strategy != "" && run.result.Audit != nil {
			run.result.Audit.Audit = strategy
		}
	}
	run.o.sendStatus(run.ctx, RoleBrain, "completed", "✅ Стратегия построена на основе анализа.", 22)

	applog(run.ctx).DebugContext(run.ctx, "planner phase start")
	run.o.sendStatus(run.ctx, RolePlanner, "running", "🧠 Планировщик Истока: построение DAG-плана...", 28)
	var err error
	run.masterPlan, err = run.o.createMasterPlan(run.ctx, run.specification, run.result.Audit)
	if err != nil {
		applog(run.ctx).ErrorContext(run.ctx, "planner failed", "error", wrapper.Wrap(err))
		_ = run.fsm.TransitionTo(domain.StateFailed, err.Error())
		run.o.sendStatus(run.ctx, RolePlanner, "error", fmt.Sprintf("❌ Ошибка планирования: %v", err), 0)

		return wrapper.Wrap(err)
	}
	applog(run.ctx).InfoContext(
		run.ctx, "planner success",
		"dagTasks", len(run.masterPlan.DAG),
		"architecture", run.masterPlan.Architecture,
	)
	run.result.MasterPlan = run.masterPlan
	run.o.sendStatus(run.ctx, RolePlanner, "completed", fmt.Sprintf("✅ DAG-план готов: %d задач", len(run.masterPlan.DAG)), 100)

	return nil
}

func (run *agentModeRun) phaseAgentFeatureApproval() error {
	const maxApprovalIterations = 5
	if run.sessionID != "" && run.o.approvalRegistry != nil {
		run.o.sendStatus(run.ctx, RolePlanner, "running", "📋 Формирование бизнес-плана для утверждения...", 33)
		businessDraft := run.o.translatePlanToBusiness(run.ctx, run.specification, run.masterPlan)

		for iteration := range maxApprovalIterations {
			run.o.approvalRegistry.Register(run.ctx, run.sessionID)
			run.o.busFromCtx(run.ctx).Publish(domain.AgentEvent{
				Kind:      domain.EventUserAction,
				Agent:     RolePlanner,
				Message:   "⏸️ Ожидание утверждения функционала...",
				DraftPlan: businessDraft,
				SessionID: run.sessionID,
				Progress:  35,
			})
			run.o.sendStatus(run.ctx, RolePlanner, "running", "⏸️ Ожидание утверждения функционала...", 35)

			decision, waitErr := run.o.approvalRegistry.WaitForApproval(run.ctx, run.sessionID)
			if waitErr != nil {
				applog(run.ctx).ErrorContext(run.ctx, "feature approval failed", "error", wrapper.Wrap(waitErr))
				_ = run.fsm.TransitionTo(domain.StateFailed, "approval wait failed: "+waitErr.Error())
				run.o.sendStatus(run.ctx, RolePlanner, "error", "🚫 Соединение потеряно — генерация остановлена", 0)

				return wrapper.Wrap(waitErr)
			}

			if decision.Approved {
				applog(run.ctx).InfoContext(
					run.ctx, "features approved",
					"iteration", iteration,
					"feedback", decision.Feedback,
				)
				run.o.sendStatus(run.ctx, RolePlanner, "completed", "✅ Функционал утверждён пользователем", 38)

				break
			}

			if strings.TrimSpace(decision.Feedback) == "" || decision.Feedback == "rejected by user" {
				_ = run.fsm.TransitionTo(domain.StateFailed, "features rejected by user")
				run.o.sendStatus(run.ctx, RolePlanner, "error", "❌ Функционал отклонён пользователем", 0)

				return ErrFeaturesRejected
			}

			applog(run.ctx).InfoContext(
				run.ctx, "feedback loop replan",
				"iteration", iteration+1,
				"feedback", decision.Feedback,
			)
			run.o.sendStatus(run.ctx, RolePlanner, "running", "🔄 Перепланирование с учётом правок...", 30)

			enrichedSpec := run.specification + "\n\n### Правки пользователя:\n" + decision.Feedback
			newPlan, planErr := run.o.createMasterPlan(run.ctx, enrichedSpec, run.result.Audit)
			if planErr != nil {
				applog(run.ctx).WarnContext(run.ctx, "replan failed, keeping plan", "error", wrapper.Wrap(planErr))
				run.o.sendStatus(run.ctx, RolePlanner, "running", "⚠️ Не удалось перепланировать — сохраняем текущий план", 33)

				continue
			}

			run.masterPlan = newPlan
			run.result.MasterPlan = run.masterPlan
			businessDraft = run.o.translatePlanToBusiness(run.ctx, enrichedSpec, run.masterPlan)
			run.specification = enrichedSpec
			applog(run.ctx).InfoContext(
				run.ctx, "replan complete",
				"iteration", iteration+1,
				"dagTasks", len(run.masterPlan.DAG),
			)
		}
	}

	return nil
}

func (run *agentModeRun) phaseAgentPostPlanFSM() error {
	err := run.fsm.ApprovePlan(domain.ApprovedPlan{
		Architecture: run.masterPlan.Architecture,
		Steps:        run.masterPlan.Steps,
		Components:   run.masterPlan.Components,
		Technologies: run.masterPlan.Technologies,
		ApprovedBy:   "user",
	})
	if err != nil {
		_ = run.fsm.TransitionTo(domain.StateFailed, "plan rejected: "+err.Error())

		return wrapper.Wrap(err)
	}
	err = run.fsm.TransitionTo(domain.StateArchitectureApproved, "user plan approved")
	if err != nil {
		return wrapper.Wrap(err)
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StatePlanning, domain.StateArchitectureApproved, "plan approved")
	run.o.busFromCtx(run.ctx).Publish(domain.AgentEvent{
		Kind: domain.EventPlan, Agent: RoleDirector,
		Message: fmt.Sprintf("%d steps, %d techs", len(run.masterPlan.Steps), len(run.masterPlan.Technologies)),
	})

	err = run.o.planner.AdvanceToStrategySynthesized(run.ctx, run.fsm, run.o.projectCtx)
	if err != nil {
		applog(run.ctx).WarnContext(run.ctx, "planner FSM gate fallback", "error", wrapper.Wrap(err))
		run.o.sendStatus(run.ctx, RolePlanner, "running", fmt.Sprintf("⚠️ Planner readiness: %v", err), 24)
		fsmErr := run.fsm.TransitionTo(domain.StateStrategySynthesized, "strategy synthesis done (fallback)")
		if fsmErr != nil {
			applog(run.ctx).WarnContext(run.ctx, "FSM strategy fallback failed", "error", wrapper.Wrap(fsmErr))
		}
	} else {
		run.o.sendStatus(run.ctx, RolePlanner, "running", "✅ Planner: readiness check passed", 26)
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateArchitectureApproved, domain.StateStrategySynthesized, "planner gate")

	err = run.fsm.TransitionTo(domain.StateDesigning, "starting design phase")
	if err != nil {
		applog(run.ctx).WarnContext(run.ctx, "FSM designing transition", "error", wrapper.Wrap(err))
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateStrategySynthesized, domain.StateDesigning, "design start")

	return nil
}

func (run *agentModeRun) phaseAgentCoding() (*GenerationResult, error) {
	err := run.fsm.TransitionTo(domain.StateCoding, "design complete, starting code generation")
	if err != nil {
		_ = run.fsm.TransitionTo(domain.StateFailed, "FSM coding gate: "+err.Error())

		return nil, wrapper.Wrap(err)
	}
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateDesigning, domain.StateCoding, "coding start")

	// Resolve media BEFORE coding. Feature gate (run.generateVideo):
	//  - false → skip Videographer entirely (fast prototype, token economy).
	//  - true  → run Videographer FIRST (sequential), then the Coder embeds the real URL.
	media := run.resolveMediaForCoder()

	run.o.sendStatus(run.ctx, RoleCoder, "running", "💻 Кодер пишет функциональный код...", 40)
	generatedCode, coderErr := run.runCoder(media)
	if coderErr != nil {
		_ = run.fsm.TransitionTo(domain.StateFailed, coderErr.Error())
		run.result.Code = map[string]string{
			"index.html": fmt.Sprintf("<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>ИСТОК</title></head><body><h1>Ошибка генерации</h1><p>%s</p><p>Повторите попытку или уточните спецификацию.</p></body></html>", coderErr.Error()),
		}
		run.result.Duration = time.Since(run.startTime)

		return run.result, wrapper.Wrap(coderErr)
	}

	if stubs := usecases.BackfillMissingImports(generatedCode); len(stubs) > 0 {
		applog(run.ctx).InfoContext(run.ctx, "backfilled import stubs", "count", len(stubs), "stubs", stubs)
	}

	for filename, content := range generatedCode {
		run.o.busFromCtx(run.ctx).PublishFile(RoleCoder, filename, content)
	}
	applog(run.ctx).InfoContext(run.ctx, "partial delivery published", "files", len(generatedCode))

	run.generatedCode = generatedCode

	return nil, ErrCodingPhaseDone
}

// resolveMediaForCoder решает медиа-контракт ДО запуска Кодера.
// GenerateVideo=false → Videographer пропускается. GenerateVideo=true → Videographer
// выполняется последовательно (до Кодера), чтобы Кодер получил реальный URL с первого
// прохода. При ошибке/таймауте видео — мягкая деградация на плейсхолдеры (Pending=true).
func (run *agentModeRun) resolveMediaForCoder() MediaContext {
	if !run.generateVideo {
		run.o.sendStatus(run.ctx, RoleVideographer, "completed", "⏭️ Видео не запрошено — пропуск (быстрый прототип)", 100)
		applog(run.ctx).InfoContext(run.ctx, "videographer skipped (GenerateVideo=false)")

		return MediaContext{Images: run.imageURLs, Pending: false, VideoRequested: false}
	}

	run.o.sendStatus(run.ctx, RoleVideographer, "running", "🎬 Монтаж промо-ролика (до кодинга)...", 38)
	video, err := run.o.uiMedia.GeneratePromoVideo(run.ctx, "ИСТОК", run.specification)
	if err != nil {
		applog(run.ctx).WarnContext(run.ctx, "videographer failed, degrading to placeholders", "error", wrapper.Wrap(err))
		run.o.sendStatus(run.ctx, RoleVideographer, "error", fmt.Sprintf("⚠️ Видео: %v — плейсхолдеры", err), 0)

		return MediaContext{Images: run.imageURLs, Pending: true, VideoRequested: true}
	}
	if video.VideoURL == "" {
		applog(run.ctx).WarnContext(run.ctx, "videographer produced no URL, degrading to placeholders")
		run.o.mu.Lock()
		// Non-sensitive metadata only — never leak the raw script/voiceover into result.Video.
		run.result.Video = fmt.Sprintf("%d scenes, %s, %s", len(video.Scenes), video.Duration, video.MusicStyle)
		run.o.mu.Unlock()
		run.o.sendStatus(run.ctx, RoleVideographer, "completed", "⚠️ Видео без URL — плейсхолдеры", 100)

		return MediaContext{Images: run.imageURLs, Pending: true, VideoRequested: true}
	}

	run.o.mu.Lock()
	run.result.Video = video.VideoURL
	run.o.mu.Unlock()
	run.o.sendStatus(run.ctx, RoleVideographer, "completed", fmt.Sprintf("✅ Промо-ролик готов: %d сцен, %s", len(video.Scenes), video.Duration), 100)

	return MediaContext{
		Images:         run.imageURLs,
		Videos:         map[string]string{"promo": video.VideoURL},
		Pending:        false,
		VideoRequested: true,
	}
}

// runCoder запускает Кодера с заданным медиа-контрактом и panic-recovery.
func (run *agentModeRun) runCoder(media MediaContext) (code map[string]string, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			applog(run.ctx).ErrorContext(run.ctx, "panic in coder", "panic", rec)
			err = wrapper.Wrapf(ErrCoderPanic, "%v", rec)
			run.o.sendStatus(run.ctx, RoleCoder, "error", fmt.Sprintf("❌ Panic: %v", rec), 0)
		}
	}()

	code, genErr := run.o.generateCodeFullStack(run.ctx, run.specification, run.masterPlan, run.result.Audit, run.manifest, run.competitorFeatures, run.imageURLs, media)
	if genErr != nil {
		run.o.sendStatus(run.ctx, RoleCoder, "error", fmt.Sprintf("❌ Ошибка кода: %v", genErr), 0)

		return nil, wrapper.Wrap(genErr)
	}
	run.o.sendStatus(run.ctx, RoleCoder, "completed", fmt.Sprintf("✅ Код сгенерирован (%d файлов), запуск валидации...", len(code)), 70)

	return code, nil
}

func (run *agentModeRun) phaseAgentVerification() (*GenerationResult, error) {
	maxRetries := run.o.autoFixMaxRetries

	// Дедлайн фазы масштабируется под число авто-исправлений: каждая попытка
	// может включать полную регенерацию Кодером (~его таймаут) плюс проверки.
	// При maxRetries == 0 остаётся компактный бюджет только на сами проверки.
	coderBudget := 10 * time.Minute
	if cfg := run.o.agents[RoleCoder]; cfg != nil && cfg.Timeout > 0 {
		coderBudget = cfg.Timeout
	}
	verificationDeadline := 90*time.Second + time.Duration(maxRetries)*(coderBudget+time.Minute)
	verifyStart := time.Now()
	gate := usecases.NewVerificationGate()
	var finalReport *usecases.VerificationReport

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if time.Since(verifyStart) > verificationDeadline {
			applog(run.ctx).WarnContext(run.ctx, "verification deadline exceeded", "deadline", verificationDeadline)
			run.o.sendStatus(run.ctx, RoleSecurity, "completed", "⏱️ Таймаут — пропущено", 100)
			run.o.sendStatus(run.ctx, RoleTester, "completed", "⏱️ Таймаут — пропущено", 100)
			run.o.sendStatus(run.ctx, RoleUIReviewer, "completed", "⏱️ Таймаут — пропущено", 100)

			break
		}

		_ = run.fsm.TransitionTo(domain.StateQualityCheck, fmt.Sprintf("verify attempt %d", attempt+1))
		run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateCoding, domain.StateQualityCheck, fmt.Sprintf("attempt %d", attempt+1))

		gate.RunTests = attempt < maxRetries

		run.o.sendStatus(run.ctx, RoleSecurity, "running", "🛡️ Security: аудит безопасности...", 80)
		run.o.sendStatus(run.ctx, RoleTester, "running", "🧪 Tester: прогон тестов...", 80)
		run.o.sendStatus(run.ctx, RoleUIReviewer, "running", "🎨 UI Reviewer: проверка UX/a11y...", 80)

		report := gate.Verify(run.ctx, run.generatedCode)
		finalReport = report
		applog(run.ctx).InfoContext(run.ctx, "verification gate attempt", "attempt", attempt+1, "summary", report.Summary)

		for _, a := range report.Approvals {
			marker := "✅"
			status := "completed"
			if !a.Approved {
				marker = "❌"
				status = "error"
			}
			applog(run.ctx).InfoContext(
				run.ctx, "verification agent result",
				"agent", a.Agent,
				"approved", a.Approved,
				"summary", a.Summary,
				"marker", marker,
			)
			switch a.Agent {
			case "security":
				run.o.sendStatus(run.ctx, RoleSecurity, status, fmt.Sprintf("%s %s", marker, a.Summary), 100)
			case "tester":
				run.o.sendStatus(run.ctx, RoleTester, status, fmt.Sprintf("%s %s", marker, a.Summary), 100)
			case "ui_reviewer":
				run.o.sendStatus(run.ctx, RoleUIReviewer, status, fmt.Sprintf("%s %s", marker, a.Summary), 100)
			}
		}

		if report.Approved {
			_ = run.fsm.TransitionTo(domain.StateSecurityCheck, "all 3 agents approved")
			run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateQualityCheck, domain.StateSecurityCheck, "verify OK")
			_ = run.fsm.TransitionTo(domain.StateVerified, "verification gate passed")
			run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateSecurityCheck, domain.StateVerified, "verified")

			break
		}

		applog(run.ctx).WarnContext(
			run.ctx, "verification gate blocked",
			"blockingAgent", report.BlockingAgent,
			"summary", report.Summary,
		)

		if attempt >= maxRetries {
			applog(run.ctx).ErrorContext(
				run.ctx, "verification max retries",
				"maxRetries", maxRetries,
				"blockingAgent", report.BlockingAgent,
			)
			_ = run.fsm.TransitionTo(domain.StateFailed,
				fmt.Sprintf("verification gate blocked by %s after %d attempts",
					report.BlockingAgent, maxRetries+1))
			run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateQualityCheck, domain.StateFailed,
				"verification blocked")

			break
		}

		retryErrorCtx := report.ForCoderContext()
		_ = run.fsm.TransitionTo(domain.StateRetryCoding,
			"auto-fix: blocked by "+report.BlockingAgent)
		run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateQualityCheck, domain.StateRetryCoding,
			"auto-fix")

		_ = run.fsm.TransitionTo(domain.StateCoding, "retry with combined error context")
		run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateRetryCoding, domain.StateCoding, "retry")

		run.o.sendStatus(run.ctx, RoleCoder, "running",
			fmt.Sprintf("🔄 Auto-fix: повторная генерация (попытка %d/%d, blocked by %s)...",
				attempt+2, maxRetries+1, report.BlockingAgent), 75)

		enrichedSpec := run.specification + "\n\n" + retryErrorCtx
		// Retry runs AFTER the videographer resolved during coding, so media is final:
		// surface the real promo video URL when present; no longer pending.
		retryMedia := MediaContext{Images: run.imageURLs, VideoRequested: run.generateVideo}
		run.o.mu.Lock()
		promoURL := run.result.Video
		run.o.mu.Unlock()
		if run.generateVideo && strings.HasPrefix(promoURL, "http") {
			retryMedia.Videos = map[string]string{"promo": promoURL}
		}
		retryCode, err := run.o.generateCodeFullStack(run.ctx, enrichedSpec, run.masterPlan, run.result.Audit,
			run.manifest, run.competitorFeatures, run.imageURLs, retryMedia)
		if err != nil {
			applog(run.ctx).WarnContext(run.ctx, "auto-fix retry failed", "attempt", attempt+1, "error", wrapper.Wrap(err))
			run.o.sendStatus(run.ctx, RoleCoder, "error", fmt.Sprintf("⚠️ Retry failed: %v", err), 0)

			break
		}
		run.generatedCode = retryCode
		run.o.sendStatus(run.ctx, RoleCoder, "completed",
			fmt.Sprintf("✅ Auto-fix код готов (%d файлов)", len(retryCode)), 78)
	}

	run.result.Code = run.generatedCode

	applog(run.ctx).DebugContext(
		run.ctx, "verification phase complete",
		"duration", time.Since(verifyStart),
		"deadline", verificationDeadline,
	)
	err := gate.CanTransitionToCompleted(finalReport)
	if err != nil {
		applog(run.ctx).WarnContext(run.ctx, "verification not fully approved, delivering anyway", "error", wrapper.Wrap(err))
		_ = run.fsm.TransitionTo(domain.StateCompleted, "delivered WITHOUT passing verification")
		run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateQualityCheck, domain.StateCompleted, "delivered without passing verification")

		run.result.Duration = time.Since(run.startTime)
		summary := "таймаут — проверка не завершена"
		if finalReport != nil {
			summary = finalReport.Summary
		}
		applog(run.ctx).WarnContext(run.ctx, "code delivered without verification", "summary", summary)
		run.o.sendStatus(run.ctx, RoleDirector, "completed",
			fmt.Sprintf("⚠️ Код доставлен за %v, но проверка качества НЕ пройдена — требуется доработка", run.result.Duration), 100)

		return run.result, nil
	}

	_ = run.fsm.TransitionTo(domain.StateCompleted, "all verification gates passed")
	run.o.busFromCtx(run.ctx).PublishFSMTransition(domain.StateVerified, domain.StateCompleted, "done")

	run.result.Duration = time.Since(run.startTime)
	applog(run.ctx).InfoContext(
		run.ctx, "fsm completed",
		"transitions", len(run.fsm.Transitions()),
		"duration", run.result.Duration,
	)

	if finalReport != nil && finalReport.TestsSkipped {
		applog(run.ctx).WarnContext(run.ctx, "verified without running tests")
		run.o.sendStatus(run.ctx, RoleDirector, "completed",
			fmt.Sprintf("🎉 Проект готов за %v (с предупреждениями: тесты не запускались)", run.result.Duration), 100)

		return run.result, nil
	}

	run.o.sendStatus(run.ctx, RoleDirector, "completed", fmt.Sprintf("🎉 Проект готов за %v", run.result.Duration), 100)

	return run.result, nil
}
