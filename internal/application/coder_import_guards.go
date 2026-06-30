package application

import (
	"regexp"
	"strings"
)

// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//  ИСТОК АГЕНТ — Import AST Guards
//  Детерминированные исправления импортов, сгенерированных LLM:
//    1. fixRemovedLucideIcons    — заменяет бренд-иконки, удалённые в lucide-react v0.400+
//    2. fixHallucinatedShadcnExports — заменяет несуществующие имена shadcn/ui
//  Обе функции мутируют переданную карту files на месте и возвращают
//  количество изменённых файлов (для телеметрии).
// ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

// removedLucideIcons — бренд-иконки, удалённые в lucide-react v0.400+.
// Значение — безопасная геометрическая замена, присутствующая во всех версиях.
var removedLucideIcons = map[string]string{
	"Github":      "Code2",
	"Gitlab":      "Code2",
	"Twitter":     "Share2",
	"Linkedin":    "Globe",
	"Facebook":    "Globe",
	"Instagram":   "Camera",
	"Youtube":     "Play",
	"Discord":     "MessageSquare",
	"Slack":       "MessageSquare",
	"Figma":       "Layers",
	"Dribbble":    "Circle",
	"Twitch":      "Play",
	"Reddit":      "MessageSquare",
	"Pinterest":   "Image",
	"Snapchat":    "Camera",
	"Tiktok":      "Music",
	"WhatsApp":    "MessageCircle",
	"Telegram":    "Send",
	"Line":        "MessageCircle",
	"WeChat":      "MessageCircle",
	"ApplePay":    "CreditCard",
	"GooglePay":   "CreditCard",
	"Paypal":      "CreditCard",
	"Stripe":      "CreditCard",
	"AmazonPay":   "CreditCard",
}

// prebuiltLucideIconRe — предкомпилированные регулярки (word-boundary) для каждой иконки.
var prebuiltLucideIconRe = func() map[string]*regexp.Regexp {
	m := make(map[string]*regexp.Regexp, len(removedLucideIcons))
	for old := range removedLucideIcons {
		m[old] = regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
	}
	return m
}()

// fixRemovedLucideIcons заменяет бренд-иконки, удалённые из lucide-react v0.400+,
// безопасными геометрическими альтернативами доступными во всех версиях библиотеки.
// Замена применяется и в import-строке, и в JSX-теле файла (слово целиком).
// Если замена вводит дубликат в import-блок — дубликат снимается.
// Возвращает количество изменённых файлов.
func fixRemovedLucideIcons(files map[string]string) int {
	count := 0
	for path, content := range files {
		if !isTSXLike(path) {
			continue
		}
		if !strings.Contains(content, "lucide-react") {
			continue
		}

		fixed := content
		for old, replacement := range removedLucideIcons {
			if !strings.Contains(fixed, old) {
				continue
			}
			fixed = prebuiltLucideIconRe[old].ReplaceAllString(fixed, replacement)
		}

		if fixed != content {
			// Remove duplicates introduced by renaming (e.g., Code2 was already imported).
			fixed = deduplicateImportSpecifiers(fixed)
			files[path] = fixed
			count++
		}
	}

	return count
}

// knownShadcnExports — канонические named exports каждого shadcn/ui компонента.
// Первый элемент — главный (primary) export, используется как fallback-замена.
var knownShadcnExports = map[string][]string{
	"button":         {"Button", "buttonVariants"},
	"card":           {"Card", "CardHeader", "CardFooter", "CardTitle", "CardDescription", "CardContent"},
	"input":          {"Input"},
	"textarea":       {"Textarea"},
	"badge":          {"Badge", "badgeVariants"},
	"dialog":         {"Dialog", "DialogContent", "DialogDescription", "DialogFooter", "DialogHeader", "DialogTitle", "DialogTrigger", "DialogClose", "DialogOverlay", "DialogPortal"},
	"tabs":           {"Tabs", "TabsContent", "TabsList", "TabsTrigger"},
	"avatar":         {"Avatar", "AvatarImage", "AvatarFallback"},
	"select":         {"Select", "SelectContent", "SelectItem", "SelectLabel", "SelectGroup", "SelectSeparator", "SelectTrigger", "SelectValue", "SelectScrollDownButton", "SelectScrollUpButton"},
	"separator":      {"Separator"},
	"label":          {"Label"},
	"checkbox":       {"Checkbox"},
	"radio-group":    {"RadioGroup", "RadioGroupItem"},
	"switch":         {"Switch"},
	"slider":         {"Slider"},
	"sheet":          {"Sheet", "SheetContent", "SheetDescription", "SheetFooter", "SheetHeader", "SheetTitle", "SheetTrigger", "SheetClose", "SheetOverlay", "SheetPortal"},
	"tooltip":        {"Tooltip", "TooltipContent", "TooltipProvider", "TooltipTrigger"},
	"progress":       {"Progress"},
	"skeleton":       {"Skeleton"},
	"scroll-area":    {"ScrollArea", "ScrollBar"},
	"dropdown-menu":  {"DropdownMenu", "DropdownMenuCheckboxItem", "DropdownMenuContent", "DropdownMenuGroup", "DropdownMenuItem", "DropdownMenuLabel", "DropdownMenuPortal", "DropdownMenuRadioGroup", "DropdownMenuRadioItem", "DropdownMenuSeparator", "DropdownMenuShortcut", "DropdownMenuSub", "DropdownMenuSubContent", "DropdownMenuSubTrigger", "DropdownMenuTrigger"},
	"popover":        {"Popover", "PopoverContent", "PopoverTrigger"},
	"alert":          {"Alert", "AlertDescription", "AlertTitle"},
	"accordion":      {"Accordion", "AccordionContent", "AccordionItem", "AccordionTrigger"},
	"table":          {"Table", "TableBody", "TableCaption", "TableCell", "TableFooter", "TableHead", "TableHeader", "TableRow"},
	"form":           {"Form", "FormControl", "FormDescription", "FormField", "FormItem", "FormLabel", "FormMessage"},
	"alert-dialog":   {"AlertDialog", "AlertDialogAction", "AlertDialogCancel", "AlertDialogContent", "AlertDialogDescription", "AlertDialogFooter", "AlertDialogHeader", "AlertDialogTitle", "AlertDialogTrigger", "AlertDialogOverlay", "AlertDialogPortal"},
	"command":        {"Command", "CommandDialog", "CommandEmpty", "CommandGroup", "CommandInput", "CommandItem", "CommandList", "CommandSeparator", "CommandShortcut"},
	"calendar":       {"Calendar"},
	"date-picker":    {"DatePicker"},
	"collapsible":    {"Collapsible", "CollapsibleContent", "CollapsibleTrigger"},
	"context-menu":   {"ContextMenu", "ContextMenuContent", "ContextMenuItem", "ContextMenuLabel", "ContextMenuSeparator", "ContextMenuTrigger"},
	"navigation-menu": {"NavigationMenu", "NavigationMenuContent", "NavigationMenuItem", "NavigationMenuLink", "NavigationMenuList", "NavigationMenuTrigger", "NavigationMenuViewport", "NavigationMenuIndicator"},
	"menubar":        {"Menubar", "MenubarContent", "MenubarItem", "MenubarMenu", "MenubarSeparator", "MenubarShortcut", "MenubarTrigger"},
	"toast":          {"Toast", "ToastAction", "ToastClose", "ToastDescription", "ToastProvider", "ToastTitle", "ToastViewport", "Toaster", "useToast"},
	"toggle":         {"Toggle", "toggleVariants"},
	"toggle-group":   {"ToggleGroup", "ToggleGroupItem"},
	"aspect-ratio":   {"AspectRatio"},
	"hover-card":     {"HoverCard", "HoverCardContent", "HoverCardTrigger"},
	"resizable":      {"ResizableHandle", "ResizablePanel", "ResizablePanelGroup"},
	"breadcrumb":     {"Breadcrumb", "BreadcrumbEllipsis", "BreadcrumbItem", "BreadcrumbLink", "BreadcrumbList", "BreadcrumbPage", "BreadcrumbSeparator"},
	"carousel":       {"Carousel", "CarouselContent", "CarouselItem", "CarouselNext", "CarouselPrevious"},
	"drawer":         {"Drawer", "DrawerClose", "DrawerContent", "DrawerDescription", "DrawerFooter", "DrawerHeader", "DrawerOverlay", "DrawerPortal", "DrawerTitle", "DrawerTrigger"},
	"input-otp":      {"InputOTP", "InputOTPGroup", "InputOTPSeparator", "InputOTPSlot"},
	"pagination":     {"Pagination", "PaginationContent", "PaginationEllipsis", "PaginationItem", "PaginationLink", "PaginationNext", "PaginationPrevious"},
	"sonner":         {"Toaster"},
	"chart":          {"ChartContainer", "ChartLegend", "ChartLegendContent", "ChartStyle", "ChartTooltip", "ChartTooltipContent"},
}

// shadcnImportRe — ищет import { ... } from '@/components/ui/COMP' (одинарные и двойные кавычки).
var shadcnImportRe = regexp.MustCompile(
	`import\s*\{([^}]+)\}\s*from\s*['"]@/components/ui/([a-zA-Z0-9_-]+)['"]`,
)

// fixHallucinatedShadcnExports детектирует и исправляет несуществующие named exports
// из shadcn/ui (например, "ModernButton" вместо "Button", "PremiumCard" вместо "Card").
//
// Алгоритм (двухпроходный):
//  1. Сканирует все shadcn import-строки → собирает пары {hallucinated → canonical}.
//  2. Применяет все замены ко всему файлу (import + JSX тело) через word-boundary regex.
//  3. Снимает дубли в import-блоках, если замена привела к повторам.
//
// Возвращает количество изменённых файлов.
func fixHallucinatedShadcnExports(files map[string]string) int {
	count := 0
	for path, content := range files {
		if !isTSXLike(path) {
			continue
		}
		if !strings.Contains(content, "@/components/ui/") {
			continue
		}

		// Шаг 1: собрать замены (сканирование без модификации).
		replacements := map[string]string{}
		for _, groups := range shadcnImportRe.FindAllStringSubmatch(content, -1) {
			if len(groups) < 3 {
				continue
			}
			compName := groups[2]
			known, ok := knownShadcnExports[compName]
			if !ok {
				continue
			}
			knownSet := make(map[string]bool, len(known))
			for _, k := range known {
				knownSet[k] = true
			}

			for _, raw := range strings.Split(groups[1], ",") {
				spec := strings.TrimSpace(raw)
				if spec == "" {
					continue
				}
				importedName := strings.Fields(spec)[0] // "Name" или "Name as Alias" → берём "Name"
				if knownSet[importedName] {
					continue // легитимный экспорт
				}
				canonical := findShadcnCanonical(importedName, known)
				if canonical != "" && canonical != importedName {
					replacements[importedName] = canonical
				}
			}
		}

		if len(replacements) == 0 {
			continue
		}

		// Шаг 2: применить все замены (word-boundary) ко всему файлу.
		fixed := content
		for old, canonical := range replacements {
			re := regexp.MustCompile(`\b` + regexp.QuoteMeta(old) + `\b`)
			fixed = re.ReplaceAllString(fixed, canonical)
		}

		// Шаг 3: снять дубли в import-блоках.
		fixed = deduplicateImportSpecifiers(fixed)

		if fixed != content {
			files[path] = fixed
			count++
		}
	}

	return count
}

// findShadcnCanonical ищет лучший канонический экспорт для hallucinated-имени.
//   - Если имя заканчивается на один из known → тот и есть canonical (e.g. "ModernButton" → "Button").
//   - Fallback: первый элемент known (primary export компонента).
func findShadcnCanonical(name string, known []string) string {
	for _, k := range known {
		if strings.HasSuffix(name, k) {
			return k
		}
	}
	if len(known) > 0 {
		return known[0]
	}
	return ""
}

// importBracesRe — ищет содержимое { ... } в import-выражениях.
var importBracesRe = regexp.MustCompile(`(import\s*\{)([^}]+)(\}\s*from)`)

// deduplicateImportSpecifiers снимает дублирующиеся named imports, возникшие после
// word-boundary замены. Сохраняет порядок первого вхождения.
func deduplicateImportSpecifiers(content string) string {
	return importBracesRe.ReplaceAllStringFunc(content, func(match string) string {
		groups := importBracesRe.FindStringSubmatch(match)
		if len(groups) < 4 {
			return match
		}
		prefix, inner, suffix := groups[1], groups[2], groups[3]

		seen := map[string]bool{}
		var deduped []string
		for _, raw := range strings.Split(inner, ",") {
			s := strings.TrimSpace(raw)
			if s == "" {
				continue
			}
			key := strings.Fields(s)[0] // первичное имя (до "as")
			if !seen[key] {
				seen[key] = true
				deduped = append(deduped, s)
			}
		}
		return prefix + " " + strings.Join(deduped, ", ") + " " + suffix
	})
}

// isTSXLike — true для файлов TypeScript/TSX/JS/JSX.
func isTSXLike(path string) bool {
	return strings.HasSuffix(path, ".tsx") ||
		strings.HasSuffix(path, ".ts") ||
		strings.HasSuffix(path, ".jsx") ||
		strings.HasSuffix(path, ".js")
}
