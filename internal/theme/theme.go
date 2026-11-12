package theme

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type Mood string

const (
	MoodDark  Mood = "dark"
	MoodLight Mood = "light"
)

type FontFace struct {
	Param string `json:"param"`
	Stack string `json:"stack"`
	Kind  string `json:"kind"`
}

type Neutrals struct {
	Crust    string `json:"crust"`
	Mantle   string `json:"mantle"`
	Base     string `json:"base"`
	Surface0 string `json:"surface0"`
	Surface1 string `json:"surface1"`
	Surface2 string `json:"surface2"`
	Overlay0 string `json:"overlay0"`
	Overlay1 string `json:"overlay1"`
	Subtext0 string `json:"subtext0"`
	Subtext1 string `json:"subtext1"`
	Text     string `json:"text"`
}

type Accents struct {
	Lavender string `json:"lavender"`
	Blue     string `json:"blue"`
	Sapphire string `json:"sapphire"`
	Sky      string `json:"sky"`
	Teal     string `json:"teal"`
	Green    string `json:"green"`
	Yellow   string `json:"yellow"`
	Peach    string `json:"peach"`
	Red      string `json:"red"`
	Mauve    string `json:"mauve"`
	Pink     string `json:"pink"`
}

type Roles struct {
	Accent  string `json:"accent"`
	Accent2 string `json:"accent2"`
	Accent3 string `json:"accent3"`
}

type TypeSpec struct {
	Display      string `json:"display"`
	Body         string `json:"body"`
	Mono         string `json:"mono"`
	Weight       int    `json:"weight"`
	Tracking     string `json:"tracking"`
	BodyTracking string `json:"bodyTracking"`
	Case         string `json:"case"`
}

type Theme struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Mood     Mood     `json:"mood"`
	Blurb    string   `json:"blurb"`
	Neutrals Neutrals `json:"neutrals"`
	Accents  Accents  `json:"accents"`
	Roles    Roles    `json:"roles"`
	Type     TypeSpec `json:"type"`
	Decor    string   `json:"decor"`
	HLJS     string   `json:"hljs"`
}

const DefaultTheme = "nord"

var ThemeIDs = []string{"midnight", "tokyo", "nord", "dracula", "gruvbox", "rosepine", "forest", "onedarkpro", "catppuccin-frappe", "catppuccin-macchiato", "catppuccin-mocha", "neon", "daylight", "arctic", "solarized", "paper", "rosequartz", "swiss", "catppuccin-latte"}

var FontIDs = []string{"inter", "spaceGrotesk", "sora", "manrope", "figtree", "outfit", "archivo", "syne", "bricolage", "workSans", "plexSans", "fraunces", "playfair", "newsreader", "lora", "plexMono", "jetbrains", "firaCode", "spaceMono", "sourceCodePro"}

var Aliases = map[string]string{
	"dark":        "nord",
	"light":       "daylight",
	"mocha":       "midnight",
	"latte":       "daylight",
	"tokyo-night": "tokyo",
	"rose-pine":   "rosepine",
	"rose-quartz": "rosequartz",
}

var Fonts = map[string]FontFace{
	"inter": {Param: "Inter:wght@300;400;500;600;700;800", Stack: "'Inter', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"spaceGrotesk": {Param: "Space+Grotesk:wght@400;500;600;700", Stack: "'Space Grotesk', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"sora": {Param: "Sora:wght@300;400;500;600;700;800", Stack: "'Sora', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"manrope": {Param: "Manrope:wght@400;500;600;700;800", Stack: "'Manrope', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"figtree": {Param: "Figtree:ital,wght@0,400;0,500;0,600;0,700;0,800;1,400", Stack: "'Figtree', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"outfit": {Param: "Outfit:wght@300;400;500;600;700;800", Stack: "'Outfit', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"archivo": {Param: "Archivo:wght@400;500;600;700;800;900", Stack: "'Archivo', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"syne": {Param: "Syne:wght@400;500;600;700;800", Stack: "'Syne', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"bricolage": {Param: "Bricolage+Grotesque:wght@400;500;600;700;800", Stack: "'Bricolage Grotesque', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"workSans": {Param: "Work+Sans:ital,wght@0,400;0,500;0,600;0,700;1,400", Stack: "'Work Sans', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"plexSans": {Param: "IBM+Plex+Sans:ital,wght@0,300;0,400;0,500;0,600;0,700;1,400", Stack: "'IBM Plex Sans', system-ui, -apple-system, 'Segoe UI', Helvetica, Arial, sans-serif", Kind: "sans"},
	"fraunces": {Param: "Fraunces:ital,opsz,wght@0,9..144,400;0,9..144,600;0,9..144,700;0,9..144,900;1,9..144,400", Stack: "'Fraunces', Georgia, 'Times New Roman', serif", Kind: "serif"},
	"playfair": {Param: "Playfair+Display:ital,wght@0,500;0,600;0,700;0,800;1,500", Stack: "'Playfair Display', Georgia, 'Times New Roman', serif", Kind: "serif"},
	"newsreader": {Param: "Newsreader:ital,opsz,wght@0,6..72,300;0,6..72,400;0,6..72,500;0,6..72,600;1,6..72,400", Stack: "'Newsreader', Georgia, 'Times New Roman', serif", Kind: "serif"},
	"lora": {Param: "Lora:ital,wght@0,400;0,500;0,600;0,700;1,400", Stack: "'Lora', Georgia, 'Times New Roman', serif", Kind: "serif"},
	"plexMono": {Param: "IBM+Plex+Mono:ital,wght@0,300;0,400;0,500;0,600;0,700;1,400", Stack: "'IBM Plex Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace", Kind: "mono"},
	"jetbrains": {Param: "JetBrains+Mono:ital,wght@0,400;0,500;0,600;0,700;0,800;1,400", Stack: "'JetBrains Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace", Kind: "mono"},
	"firaCode": {Param: "Fira+Code:wght@400;500;600;700", Stack: "'Fira Code', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace", Kind: "mono"},
	"spaceMono": {Param: "Space+Mono:ital,wght@0,400;0,700;1,400", Stack: "'Space Mono', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace", Kind: "mono"},
	"sourceCodePro": {Param: "Source+Code+Pro:ital,wght@0,400;0,500;0,600;0,700;1,400", Stack: "'Source Code Pro', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, Consolas, monospace", Kind: "mono"},
}

var Themes = map[string]Theme{
	"midnight": {
		ID: "midnight",
		Label: "midnight",
		Mood: Mood("dark"),
		Blurb: "Catppuccin Mocha. Violet on deep indigo, drifting orbs.",
		Neutrals: Neutrals{
			Crust: "#11111b",
			Mantle: "#181825",
			Base: "#1e1e2e",
			Surface0: "#313244",
			Surface1: "#45475a",
			Surface2: "#585b70",
			Overlay0: "#6c7086",
			Overlay1: "#7f849c",
			Subtext0: "#a6adc8",
			Subtext1: "#bac2de",
			Text: "#cdd6f4",
		},
		Accents: Accents{
			Lavender: "#b4befe",
			Blue: "#89b4fa",
			Sapphire: "#74c7ec",
			Sky: "#89dceb",
			Teal: "#94e2d5",
			Green: "#a6e3a1",
			Yellow: "#f9e2af",
			Peach: "#fab387",
			Red: "#f38ba8",
			Mauve: "#cba6f7",
			Pink: "#f5c2e7",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "blue",
			Accent3: "sky",
		},
		Type: TypeSpec{
			Display: "spaceGrotesk",
			Body: "inter",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.025em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "orbs",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/tokyo-night-dark.min.css",
	},
	"tokyo": {
		ID: "tokyo",
		Label: "tokyo night",
		Mood: Mood("dark"),
		Blurb: "Neon cyan over a wireframe grid, lifted from a rainy skyline.",
		Neutrals: Neutrals{
			Crust: "#16161e",
			Mantle: "#1a1b26",
			Base: "#1f2335",
			Surface0: "#292e42",
			Surface1: "#3b4261",
			Surface2: "#545c7e",
			Overlay0: "#626c96",
			Overlay1: "#7c86ab",
			Subtext0: "#a5b0d6",
			Subtext1: "#b7c0e0",
			Text: "#c8d3f5",
		},
		Accents: Accents{
			Lavender: "#a9b1ff",
			Blue: "#7aa2f7",
			Sapphire: "#2ac3de",
			Sky: "#7dcfff",
			Teal: "#73daca",
			Green: "#9ece6a",
			Yellow: "#e0af68",
			Peach: "#ff9e64",
			Red: "#f7768e",
			Mauve: "#bb9af7",
			Pink: "#ff7eb6",
		},
		Roles: Roles{
			Accent: "sapphire",
			Accent2: "mauve",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "sora",
			Body: "inter",
			Mono: "jetbrains",
			Weight: 700,
			Tracking: "-0.03em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "grid",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/tokyo-night-dark.min.css",
	},
	"nord": {
		ID: "nord",
		Label: "nord",
		Mood: Mood("dark"),
		Blurb: "Arctic frost blue on polar slate, with slow contour waves.",
		Neutrals: Neutrals{
			Crust: "#232831",
			Mantle: "#2e3440",
			Base: "#353c4a",
			Surface0: "#3b4252",
			Surface1: "#434c5e",
			Surface2: "#4c566a",
			Overlay0: "#707d92",
			Overlay1: "#8e9aae",
			Subtext0: "#c3ccda",
			Subtext1: "#d8dee9",
			Text: "#eceff4",
		},
		Accents: Accents{
			Lavender: "#c9a9c8",
			Blue: "#81a1c1",
			Sapphire: "#88c0d0",
			Sky: "#a3d4e0",
			Teal: "#8fbcbb",
			Green: "#a3be8c",
			Yellow: "#ebcb8b",
			Peach: "#d08770",
			Red: "#bf616a",
			Mauve: "#b48ead",
			Pink: "#d3a3c6",
		},
		Roles: Roles{
			Accent: "sapphire",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "jetbrains",
			Body: "jetbrains",
			Mono: "jetbrains",
			Weight: 700,
			Tracking: "-0.025em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "waves",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/nord.min.css",
	},
	"dracula": {
		ID: "dracula",
		Label: "dracula",
		Mood: Mood("dark"),
		Blurb: "Purple and hot pink over charcoal, lit by a gradient mesh.",
		Neutrals: Neutrals{
			Crust: "#1a1b23",
			Mantle: "#21222c",
			Base: "#282a36",
			Surface0: "#343746",
			Surface1: "#44475a",
			Surface2: "#565a70",
			Overlay0: "#6272a4",
			Overlay1: "#7b85b6",
			Subtext0: "#c8cbe0",
			Subtext1: "#e2e4f0",
			Text: "#f8f8f2",
		},
		Accents: Accents{
			Lavender: "#d6bcff",
			Blue: "#79b8ff",
			Sapphire: "#8be9fd",
			Sky: "#a4f0fd",
			Teal: "#66e0c8",
			Green: "#50fa7b",
			Yellow: "#f1fa8c",
			Peach: "#ffb86c",
			Red: "#ff5555",
			Mauve: "#bd93f9",
			Pink: "#ff79c6",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "pink",
			Accent3: "sapphire",
		},
		Type: TypeSpec{
			Display: "syne",
			Body: "figtree",
			Mono: "firaCode",
			Weight: 800,
			Tracking: "-0.02em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "mesh",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/dracula.min.css",
	},
	"gruvbox": {
		ID: "gruvbox",
		Label: "gruvbox",
		Mood: Mood("dark"),
		Blurb: "Warm amber and moss on retro brown, hatched like graph paper.",
		Neutrals: Neutrals{
			Crust: "#1b1e1f",
			Mantle: "#232728",
			Base: "#282828",
			Surface0: "#32302f",
			Surface1: "#3c3836",
			Surface2: "#504945",
			Overlay0: "#7c6f64",
			Overlay1: "#928374",
			Subtext0: "#bdae93",
			Subtext1: "#d5c4a1",
			Text: "#ebdbb2",
		},
		Accents: Accents{
			Lavender: "#d3b8c8",
			Blue: "#83a598",
			Sapphire: "#7daea3",
			Sky: "#8ec07c",
			Teal: "#89b482",
			Green: "#b8bb26",
			Yellow: "#fabd2f",
			Peach: "#fe8019",
			Red: "#fb4934",
			Mauve: "#d3869b",
			Pink: "#e2a3b4",
		},
		Roles: Roles{
			Accent: "yellow",
			Accent2: "peach",
			Accent3: "green",
		},
		Type: TypeSpec{
			Display: "bricolage",
			Body: "workSans",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.02em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "topo",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/gruvbox-dark-medium.min.css",
	},
	"rosepine": {
		ID: "rosepine",
		Label: "rosé pine",
		Mood: Mood("dark"),
		Blurb: "Muted iris and gold on plum, under a slow aurora.",
		Neutrals: Neutrals{
			Crust: "#14121f",
			Mantle: "#191724",
			Base: "#1f1d2e",
			Surface0: "#26233a",
			Surface1: "#403d52",
			Surface2: "#524f67",
			Overlay0: "#6e6a86",
			Overlay1: "#817c9c",
			Subtext0: "#b6b2d0",
			Subtext1: "#cdc9e6",
			Text: "#e0def4",
		},
		Accents: Accents{
			Lavender: "#d5c4ee",
			Blue: "#6a9fb5",
			Sapphire: "#6ba8bf",
			Sky: "#9ccfd8",
			Teal: "#86cfc4",
			Green: "#9ccfa8",
			Yellow: "#f6c177",
			Peach: "#f0a882",
			Red: "#eb6f92",
			Mauve: "#c4a7e7",
			Pink: "#ebbcba",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "sky",
			Accent3: "yellow",
		},
		Type: TypeSpec{
			Display: "fraunces",
			Body: "newsreader",
			Mono: "plexMono",
			Weight: 600,
			Tracking: "-0.02em",
			BodyTracking: "0.004em",
			Case: "none",
		},
		Decor: "aurora",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/ros-pine.min.css",
	},
	"forest": {
		ID: "forest",
		Label: "forest",
		Mood: Mood("dark"),
		Blurb: "Everforest sage on deep pine, rippling in concentric rings.",
		Neutrals: Neutrals{
			Crust: "#232a2e",
			Mantle: "#2d353b",
			Base: "#343f44",
			Surface0: "#3d484d",
			Surface1: "#475258",
			Surface2: "#4f585e",
			Overlay0: "#7a8478",
			Overlay1: "#9da9a0",
			Subtext0: "#c2c9bd",
			Subtext1: "#d3c6aa",
			Text: "#e5dfd2",
		},
		Accents: Accents{
			Lavender: "#d8b9c8",
			Blue: "#7fbbb3",
			Sapphire: "#6fb5ac",
			Sky: "#9ed3c8",
			Teal: "#83c092",
			Green: "#a7c080",
			Yellow: "#dbbc7f",
			Peach: "#e69875",
			Red: "#e67e80",
			Mauve: "#d699b6",
			Pink: "#e3aec3",
		},
		Roles: Roles{
			Accent: "green",
			Accent2: "sapphire",
			Accent3: "yellow",
		},
		Type: TypeSpec{
			Display: "outfit",
			Body: "figtree",
			Mono: "jetbrains",
			Weight: 600,
			Tracking: "-0.025em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "rings",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css",
	},
	"onedarkpro": {
		ID: "onedarkpro",
		Label: "one dark pro",
		Mood: Mood("dark"),
		Blurb: "Atom's One Dark Pro. Muted blue-grey with cyan and violet accents, on charcoal.",
		Neutrals: Neutrals{
			Crust: "#181a1f",
			Mantle: "#21252b",
			Base: "#282c34",
			Surface0: "#2c313c",
			Surface1: "#3e4452",
			Surface2: "#495162",
			Overlay0: "#4e5666",
			Overlay1: "#5a6375",
			Subtext0: "#5c6370",
			Subtext1: "#abb2bf",
			Text: "#d7dae0",
		},
		Accents: Accents{
			Lavender: "#c162de",
			Blue: "#61afef",
			Sapphire: "#528bff",
			Sky: "#4dc4ff",
			Teal: "#56b6c2",
			Green: "#98c379",
			Yellow: "#e5c07b",
			Peach: "#d19a66",
			Red: "#e06c75",
			Mauve: "#c678dd",
			Pink: "#de73ff",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "manrope",
			Body: "inter",
			Mono: "jetbrains",
			Weight: 700,
			Tracking: "-0.03em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "grid",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-dark.min.css",
	},
	"catppuccin-frappe": {
		ID: "catppuccin-frappe",
		Label: "catppuccin frappé",
		Mood: Mood("dark"),
		Blurb: "Official Catppuccin Frappé. Muted mauve on soft slate blue.",
		Neutrals: Neutrals{
			Crust: "#232634",
			Mantle: "#292c3c",
			Base: "#303446",
			Surface0: "#414559",
			Surface1: "#51576d",
			Surface2: "#626880",
			Overlay0: "#737994",
			Overlay1: "#838ba7",
			Subtext0: "#a5adce",
			Subtext1: "#b5bfe2",
			Text: "#c6d0f5",
		},
		Accents: Accents{
			Lavender: "#babbf1",
			Blue: "#8caaee",
			Sapphire: "#85c1dc",
			Sky: "#99d1db",
			Teal: "#81c8be",
			Green: "#a6d189",
			Yellow: "#e5c890",
			Peach: "#ef9f76",
			Red: "#e78284",
			Mauve: "#ca9ee6",
			Pink: "#f4b8e4",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "sky",
			Accent3: "green",
		},
		Type: TypeSpec{
			Display: "figtree",
			Body: "inter",
			Mono: "jetbrains",
			Weight: 600,
			Tracking: "-0.02em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "dots",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/nord.min.css",
	},
	"catppuccin-macchiato": {
		ID: "catppuccin-macchiato",
		Label: "catppuccin macchiato",
		Mood: Mood("dark"),
		Blurb: "Official Catppuccin Macchiato. Balanced mauve on deep indigo.",
		Neutrals: Neutrals{
			Crust: "#181926",
			Mantle: "#1e2030",
			Base: "#24273a",
			Surface0: "#363a4f",
			Surface1: "#494d64",
			Surface2: "#5b6078",
			Overlay0: "#6e738d",
			Overlay1: "#8087a2",
			Subtext0: "#a5adcb",
			Subtext1: "#b8c0e0",
			Text: "#cad3f5",
		},
		Accents: Accents{
			Lavender: "#b7bdf8",
			Blue: "#8aadf4",
			Sapphire: "#7dc4e4",
			Sky: "#91d7e3",
			Teal: "#8bd5ca",
			Green: "#a6da95",
			Yellow: "#eed49f",
			Peach: "#f5a97f",
			Red: "#ed8796",
			Mauve: "#c6a0f6",
			Pink: "#f5bde6",
		},
		Roles: Roles{
			Accent: "pink",
			Accent2: "blue",
			Accent3: "sapphire",
		},
		Type: TypeSpec{
			Display: "outfit",
			Body: "plexSans",
			Mono: "firaCode",
			Weight: 700,
			Tracking: "-0.03em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "rings",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/dracula.min.css",
	},
	"catppuccin-mocha": {
		ID: "catppuccin-mocha",
		Label: "catppuccin mocha",
		Mood: Mood("dark"),
		Blurb: "Official Catppuccin Mocha. Full-strength mauve on the darkest base.",
		Neutrals: Neutrals{
			Crust: "#11111b",
			Mantle: "#181825",
			Base: "#1e1e2e",
			Surface0: "#313244",
			Surface1: "#45475a",
			Surface2: "#585b70",
			Overlay0: "#6c7086",
			Overlay1: "#7f849c",
			Subtext0: "#a6adc8",
			Subtext1: "#bac2de",
			Text: "#cdd6f4",
		},
		Accents: Accents{
			Lavender: "#b4befe",
			Blue: "#89b4fa",
			Sapphire: "#74c7ec",
			Sky: "#89dceb",
			Teal: "#94e2d5",
			Green: "#a6e3a1",
			Yellow: "#f9e2af",
			Peach: "#fab387",
			Red: "#f38ba8",
			Mauve: "#cba6f7",
			Pink: "#f5c2e7",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "manrope",
			Body: "inter",
			Mono: "jetbrains",
			Weight: 700,
			Tracking: "-0.03em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "waves",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/tokyo-night-dark.min.css",
	},
	"neon": {
		ID: "neon",
		Label: "neon",
		Mood: Mood("dark"),
		Blurb: "Electric cyan and magenta on true black, raked by light beams.",
		Neutrals: Neutrals{
			Crust: "#050509",
			Mantle: "#0a0a12",
			Base: "#0e0e1a",
			Surface0: "#16162a",
			Surface1: "#22223c",
			Surface2: "#2e2e50",
			Overlay0: "#4f4f78",
			Overlay1: "#6a6a96",
			Subtext0: "#a8a8cc",
			Subtext1: "#c8c8e4",
			Text: "#f0f0ff",
		},
		Accents: Accents{
			Lavender: "#b39dff",
			Blue: "#4d9fff",
			Sapphire: "#22d3ee",
			Sky: "#38e8ff",
			Teal: "#2dd4bf",
			Green: "#4ade80",
			Yellow: "#fde047",
			Peach: "#fb923c",
			Red: "#fb5c7d",
			Mauve: "#d946ef",
			Pink: "#ff5cc8",
		},
		Roles: Roles{
			Accent: "sapphire",
			Accent2: "pink",
			Accent3: "sky",
		},
		Type: TypeSpec{
			Display: "spaceGrotesk",
			Body: "spaceGrotesk",
			Mono: "spaceMono",
			Weight: 700,
			Tracking: "0.04em",
			BodyTracking: "0",
			Case: "uppercase",
		},
		Decor: "beams",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/night-owl.min.css",
	},
	"daylight": {
		ID: "daylight",
		Label: "daylight",
		Mood: Mood("light"),
		Blurb: "Catppuccin Latte, contrast-tuned for a projector. Dot matrix.",
		Neutrals: Neutrals{
			Crust: "#ebebeb",
			Mantle: "#f5f5f5",
			Base: "#fafaf8",
			Surface0: "#ccd0da",
			Surface1: "#9ca0b0",
			Surface2: "#8c8fa1",
			Overlay0: "#6c6f85",
			Overlay1: "#5c5f77",
			Subtext0: "#4c4f69",
			Subtext1: "#3a3c52",
			Text: "#1e2030",
		},
		Accents: Accents{
			Lavender: "#7287fd",
			Blue: "#1e66f5",
			Sapphire: "#209fb5",
			Sky: "#04a5e5",
			Teal: "#179299",
			Green: "#40a02b",
			Yellow: "#df8e1d",
			Peach: "#fe640b",
			Red: "#d20f39",
			Mauve: "#8839ef",
			Pink: "#ea76cb",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "spaceGrotesk",
			Body: "inter",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.025em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "dots",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-light.min.css",
	},
	"arctic": {
		ID: "arctic",
		Label: "arctic",
		Mood: Mood("light"),
		Blurb: "Nord inverted. Frost blue on cool paper, with contour waves.",
		Neutrals: Neutrals{
			Crust: "#e3e8ef",
			Mantle: "#eceff4",
			Base: "#f7f9fc",
			Surface0: "#dbe1ea",
			Surface1: "#c2cad6",
			Surface2: "#a3adbd",
			Overlay0: "#7b8797",
			Overlay1: "#5f6b7c",
			Subtext0: "#445060",
			Subtext1: "#364150",
			Text: "#2e3440",
		},
		Accents: Accents{
			Lavender: "#9c6f96",
			Blue: "#4a7ba7",
			Sapphire: "#2f7d94",
			Sky: "#3d8fa8",
			Teal: "#3b8b8a",
			Green: "#5d8a45",
			Yellow: "#b08834",
			Peach: "#c06a4a",
			Red: "#b0505a",
			Mauve: "#96628f",
			Pink: "#a86a94",
		},
		Roles: Roles{
			Accent: "sapphire",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "manrope",
			Body: "plexSans",
			Mono: "jetbrains",
			Weight: 800,
			Tracking: "-0.035em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "waves",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-light.min.css",
	},
	"solarized": {
		ID: "solarized",
		Label: "solarized",
		Mood: Mood("light"),
		Blurb: "The classic low-glare cream, paired with Lora for long prose.",
		Neutrals: Neutrals{
			Crust: "#eee8d5",
			Mantle: "#f5efdc",
			Base: "#fdf6e3",
			Surface0: "#e3ddc8",
			Surface1: "#c9c3ad",
			Surface2: "#a8a48e",
			Overlay0: "#839496",
			Overlay1: "#657b83",
			Subtext0: "#586e75",
			Subtext1: "#3f5b62",
			Text: "#073642",
		},
		Accents: Accents{
			Lavender: "#8a8fd0",
			Blue: "#268bd2",
			Sapphire: "#1f8fa8",
			Sky: "#2aa198",
			Teal: "#21958c",
			Green: "#859900",
			Yellow: "#b58900",
			Peach: "#cb4b16",
			Red: "#dc322f",
			Mauve: "#6c71c4",
			Pink: "#d33682",
		},
		Roles: Roles{
			Accent: "blue",
			Accent2: "mauve",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "outfit",
			Body: "lora",
			Mono: "sourceCodePro",
			Weight: 600,
			Tracking: "-0.02em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "topo",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/solarized-light.min.css",
	},
	"paper": {
		ID: "paper",
		Label: "paper",
		Mood: Mood("light"),
		Blurb: "Crimson serif on warm cream. Editorial, print-first, very legible.",
		Neutrals: Neutrals{
			Crust: "#efe8dc",
			Mantle: "#f7f2e9",
			Base: "#fdfbf5",
			Surface0: "#e4dccc",
			Surface1: "#cec4b0",
			Surface2: "#b0a48d",
			Overlay0: "#8b8069",
			Overlay1: "#6e6553",
			Subtext0: "#4d4535",
			Subtext1: "#38311f",
			Text: "#211c11",
		},
		Accents: Accents{
			Lavender: "#6f5ba8",
			Blue: "#2d5fa8",
			Sapphire: "#1f7f96",
			Sky: "#10788f",
			Teal: "#1c7a6b",
			Green: "#3f7a2e",
			Yellow: "#9a6c12",
			Peach: "#b35311",
			Red: "#a81f2c",
			Mauve: "#7a3f9c",
			Pink: "#a83f77",
		},
		Roles: Roles{
			Accent: "red",
			Accent2: "sapphire",
			Accent3: "yellow",
		},
		Type: TypeSpec{
			Display: "fraunces",
			Body: "newsreader",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.02em",
			BodyTracking: "0.003em",
			Case: "none",
		},
		Decor: "rings",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-light.min.css",
	},
	"rosequartz": {
		ID: "rosequartz",
		Label: "rose quartz",
		Mood: Mood("light"),
		Blurb: "Rosé Pine Dawn. Blush and iris on linen, with soft orbs.",
		Neutrals: Neutrals{
			Crust: "#f2e9e1",
			Mantle: "#faf4ed",
			Base: "#fffaf3",
			Surface0: "#ece0d8",
			Surface1: "#d8ccc6",
			Surface2: "#bdb0ae",
			Overlay0: "#9893a5",
			Overlay1: "#7d7791",
			Subtext0: "#625b7d",
			Subtext1: "#55506e",
			Text: "#423d5c",
		},
		Accents: Accents{
			Lavender: "#a68fc0",
			Blue: "#286983",
			Sapphire: "#2f7d92",
			Sky: "#56949f",
			Teal: "#4a9188",
			Green: "#5a8f52",
			Yellow: "#ea9d34",
			Peach: "#d7827e",
			Red: "#b4637a",
			Mauve: "#907aa9",
			Pink: "#c4759a",
		},
		Roles: Roles{
			Accent: "red",
			Accent2: "mauve",
			Accent3: "sky",
		},
		Type: TypeSpec{
			Display: "playfair",
			Body: "figtree",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.02em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "orbs",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/base16/ros-pine-dawn.min.css",
	},
	"swiss": {
		ID: "swiss",
		Label: "swiss",
		Mood: Mood("light"),
		Blurb: "Black on white, one red. Heavy grotesk, tight tracking, hard grid.",
		Neutrals: Neutrals{
			Crust: "#f0f0f0",
			Mantle: "#f8f8f8",
			Base: "#ffffff",
			Surface0: "#dcdcdc",
			Surface1: "#bebebe",
			Surface2: "#9a9a9a",
			Overlay0: "#767676",
			Overlay1: "#5a5a5a",
			Subtext0: "#3a3a3a",
			Subtext1: "#222222",
			Text: "#0a0a0a",
		},
		Accents: Accents{
			Lavender: "#4b4b8f",
			Blue: "#1a4fd6",
			Sapphire: "#0a6e8a",
			Sky: "#0e7490",
			Teal: "#0f766e",
			Green: "#15803d",
			Yellow: "#a16207",
			Peach: "#c2410c",
			Red: "#e01b24",
			Mauve: "#5b21b6",
			Pink: "#be185d",
		},
		Roles: Roles{
			Accent: "red",
			Accent2: "text",
			Accent3: "overlay1",
		},
		Type: TypeSpec{
			Display: "archivo",
			Body: "inter",
			Mono: "plexMono",
			Weight: 900,
			Tracking: "-0.045em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "grid",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/github.min.css",
	},
	"catppuccin-latte": {
		ID: "catppuccin-latte",
		Label: "catppuccin latte",
		Mood: Mood("light"),
		Blurb: "Official Catppuccin Latte. Warm mauve on soft cream.",
		Neutrals: Neutrals{
			Crust: "#dce0e8",
			Mantle: "#e6e9ef",
			Base: "#eff1f5",
			Surface0: "#ccd0da",
			Surface1: "#bcc0cc",
			Surface2: "#acb0be",
			Overlay0: "#9ca0b0",
			Overlay1: "#8c8fa1",
			Subtext0: "#6c6f85",
			Subtext1: "#5c5f77",
			Text: "#4c4f69",
		},
		Accents: Accents{
			Lavender: "#7287fd",
			Blue: "#1e66f5",
			Sapphire: "#209fb5",
			Sky: "#04a5e5",
			Teal: "#179299",
			Green: "#40a02b",
			Yellow: "#df8e1d",
			Peach: "#fe640b",
			Red: "#d20f39",
			Mauve: "#8839ef",
			Pink: "#ea76cb",
		},
		Roles: Roles{
			Accent: "mauve",
			Accent2: "blue",
			Accent3: "teal",
		},
		Type: TypeSpec{
			Display: "spaceGrotesk",
			Body: "figtree",
			Mono: "plexMono",
			Weight: 700,
			Tracking: "-0.025em",
			BodyTracking: "0",
			Case: "none",
		},
		Decor: "aurora",
		HLJS: "https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/styles/atom-one-light.min.css",
	},
}

func FontName(key string) string {
	f, ok := Fonts[key]
	if !ok {
		return key
	}
	parts := strings.Split(f.Param, ":")
	return strings.ReplaceAll(parts[0], "+", " ")
}

func FindFont(input string) (string, bool) {
	key := strings.TrimSpace(input)
	if key == "" {
		return "", false
	}
	if _, ok := Fonts[key]; ok {
		return key, true
	}
	wanted := strings.ToLower(key)
	wanted = strings.ReplaceAll(wanted, " ", "")
	wanted = strings.ReplaceAll(wanted, "-", "")
	wanted = strings.ReplaceAll(wanted, "_", "")
	for _, id := range FontIDs {
		if strings.ToLower(id) == wanted {
			return id, true
		}
		fn := strings.ToLower(FontName(id))
		fn = strings.ReplaceAll(fn, " ", "")
		if fn == wanted {
			return id, true
		}
	}
	return "", false
}

type FontSummary struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Kind  string `json:"kind"`
	Stack string `json:"stack"`
}

func FontSummaries() []FontSummary {
	res := make([]FontSummary, len(FontIDs))
	for i, id := range FontIDs {
		f := Fonts[id]
		res[i] = FontSummary{
			ID:    id,
			Name:  FontName(id),
			Kind:  f.Kind,
			Stack: f.Stack,
		}
	}
	return res
}

func FontListing() []string {
	maxPad := 0
	for _, id := range FontIDs {
		if len(id) > maxPad {
			maxPad = len(id)
		}
	}
	lines := make([]string, len(FontIDs))
	for i, id := range FontIDs {
		f := Fonts[id]
		lines[i] = fmt.Sprintf("%-*s  %-5s  %s", maxPad, id, f.Kind, FontName(id))
	}
	return lines
}

func FindTheme(input string) (string, bool) {
	key := strings.ToLower(strings.TrimSpace(input))
	if key == "" {
		return "", false
	}
	if _, ok := Themes[key]; ok {
		return key, true
	}
	if alias, ok := Aliases[key]; ok {
		return alias, true
	}
	return "", false
}

func ResolveThemeName(input string) string {
	if t, ok := FindTheme(input); ok {
		return t
	}
	return DefaultTheme
}

type ThemeColorsSummary struct {
	Crust    string `json:"crust"`
	Mantle   string `json:"mantle"`
	Base     string `json:"base"`
	Surface0 string `json:"surface0"`
	Overlay1 string `json:"overlay1"`
	Subtext1 string `json:"subtext1"`
	Text     string `json:"text"`
	Accent   string `json:"accent"`
	Accent2  string `json:"accent2"`
	Accent3  string `json:"accent3"`
}

type ThemeFontsSummary struct {
	Display string `json:"display"`
	Body    string `json:"body"`
	Mono    string `json:"mono"`
}

type ThemeSummary struct {
	ID     string              `json:"id"`
	Label  string              `json:"label"`
	Mood   Mood                `json:"mood"`
	Blurb  string              `json:"blurb"`
	Decor  string              `json:"decor"`
	Colors ThemeColorsSummary  `json:"colors"`
	Fonts  ThemeFontsSummary   `json:"fonts"`
}

func (t Theme) ColorByKey(key string) string {
	switch key {
	case "crust": return t.Neutrals.Crust
	case "mantle": return t.Neutrals.Mantle
	case "base": return t.Neutrals.Base
	case "surface0": return t.Neutrals.Surface0
	case "surface1": return t.Neutrals.Surface1
	case "surface2": return t.Neutrals.Surface2
	case "overlay0": return t.Neutrals.Overlay0
	case "overlay1": return t.Neutrals.Overlay1
	case "subtext0": return t.Neutrals.Subtext0
	case "subtext1": return t.Neutrals.Subtext1
	case "text": return t.Neutrals.Text
	case "lavender": return t.Accents.Lavender
	case "blue": return t.Accents.Blue
	case "sapphire": return t.Accents.Sapphire
	case "sky": return t.Accents.Sky
	case "teal": return t.Accents.Teal
	case "green": return t.Accents.Green
	case "yellow": return t.Accents.Yellow
	case "peach": return t.Accents.Peach
	case "red": return t.Accents.Red
	case "mauve": return t.Accents.Mauve
	case "pink": return t.Accents.Pink
	default: return ""
	}
}

func (t Theme) RoleColor(which string) string {
	switch which {
	case "accent": return t.ColorByKey(t.Roles.Accent)
	case "accent2": return t.ColorByKey(t.Roles.Accent2)
	case "accent3": return t.ColorByKey(t.Roles.Accent3)
	default: return ""
	}
}

func ThemeSummaries() []ThemeSummary {
	res := make([]ThemeSummary, len(ThemeIDs))
	for i, id := range ThemeIDs {
		t := Themes[id]
		res[i] = ThemeSummary{
			ID:    id,
			Label: t.Label,
			Mood:  t.Mood,
			Blurb: t.Blurb,
			Decor: t.Decor,
			Colors: ThemeColorsSummary{
				Crust:    t.Neutrals.Crust,
				Mantle:   t.Neutrals.Mantle,
				Base:     t.Neutrals.Base,
				Surface0: t.Neutrals.Surface0,
				Overlay1: t.Neutrals.Overlay1,
				Subtext1: t.Neutrals.Subtext1,
				Text:     t.Neutrals.Text,
				Accent:   t.RoleColor("accent"),
				Accent2:  t.RoleColor("accent2"),
				Accent3:  t.RoleColor("accent3"),
			},
			Fonts: ThemeFontsSummary{
				Display: Fonts[t.Type.Display].Stack,
				Body:    Fonts[t.Type.Body].Stack,
				Mono:    Fonts[t.Type.Mono].Stack,
			},
		}
	}
	return res
}

func ThemeListing() []string {
	maxPad := 0
	for _, id := range ThemeIDs {
		if len(id) > maxPad {
			maxPad = len(id)
		}
	}
	lines := make([]string, len(ThemeIDs))
	for i, id := range ThemeIDs {
		t := Themes[id]
		moodStr := "dark "
		if t.Mood == MoodLight {
			moodStr = "light"
		}
		lines[i] = fmt.Sprintf("%-*s  %s  %s", maxPad, id, moodStr, t.Blurb)
	}
	return lines
}

func hexChannels(hex string) (int, int, int) {
	h := strings.TrimPrefix(hex, "#")
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	n, err := strconv.ParseUint(h, 16, 32)
	if err != nil {
		return 0, 0, 0
	}
	return int((n >> 16) & 255), int((n >> 8) & 255), int(n & 255)
}

func channelsStr(hex string) string {
	r, g, b := hexChannels(hex)
	return fmt.Sprintf("%d, %d, %d", r, g, b)
}

func alphaStr(hex string, a float64) string {
	return fmt.Sprintf("rgba(%s, %v)", channelsStr(hex), a)
}

func ThemeVars(t Theme) string {
	n := t.Neutrals
	a := t.Accents
	dark := t.Mood == MoodDark
	accent := t.RoleColor("accent")
	accent2 := t.RoleColor("accent2")
	accent3 := t.RoleColor("accent3")

	sh := channelsStr(n.Text)
	if dark {
		sh = "0, 0, 0"
	}
	s := func(y, blur int, op float64) string {
		actualOp := op
		if !dark {
			actualOp = op * 0.42
		}
		return fmt.Sprintf("0 %dpx %dpx rgba(%s, %v)", y, blur, sh, actualOp)
	}

	accentSoftOp := 0.09
	accentLineOp := 0.34
	glowOp := 0.22
	selBgOp := 0.18
	surfaceSoftOp := 0.42
	crustOverlayOp := 0.88
	scrim := alphaStr(n.Text, 0.26)
	hairlineOp := 0.11
	decorLineOp := 0.07
	decorDotOp := 0.1
	decorGlow1Op := 0.15
	decorGlow2Op := 0.12
	decorGlow3Op := 0.1

	if dark {
		accentSoftOp = 0.1
		accentLineOp = 0.42
		glowOp = 0.34
		selBgOp = 0.22
		surfaceSoftOp = 0.34
		crustOverlayOp = 0.84
		scrim = "rgba(0, 0, 0, 0.5)"
		hairlineOp = 0.09
		decorLineOp = 0.05
		decorDotOp = 0.085
		decorGlow1Op = 0.2
		decorGlow2Op = 0.16
		decorGlow3Op = 0.13
	}

	lines := [][2]string{
		{"crust", n.Crust}, {"mantle", n.Mantle}, {"base", n.Base},
		{"surface0", n.Surface0}, {"surface1", n.Surface1}, {"surface2", n.Surface2},
		{"overlay0", n.Overlay0}, {"overlay1", n.Overlay1},
		{"subtext0", n.Subtext0}, {"subtext1", n.Subtext1}, {"text", n.Text},

		{"lavender", a.Lavender}, {"blue", a.Blue}, {"sapphire", a.Sapphire},
		{"sky", a.Sky}, {"teal", a.Teal}, {"green", a.Green}, {"yellow", a.Yellow},
		{"peach", a.Peach}, {"red", a.Red}, {"mauve", a.Mauve}, {"pink", a.Pink},

		{"accent", accent},
		{"accent-2", accent2},
		{"accent-3", accent3},
		{"accent-soft", alphaStr(accent, accentSoftOp)},
		{"accent-line", alphaStr(accent, accentLineOp)},
		{"glow", alphaStr(accent, glowOp)},
		{"gradient", fmt.Sprintf("linear-gradient(115deg, %s, %s 58%%, %s)", accent, accent2, accent3)},
		{"accent-fade", fmt.Sprintf("linear-gradient(90deg, %s, %s 62%%, transparent)", accent, alphaStr(accent, 0.45))},
		{"selection-bg", alphaStr(accent, selBgOp)},
		{"selection-text", "inherit"},

		{"surface-soft", alphaStr(n.Surface0, surfaceSoftOp)},
		{"crust-overlay", alphaStr(n.Crust, crustOverlayOp)},
		{"scrim", scrim},
		{"hairline", alphaStr(n.Text, hairlineOp)},

		{"shadow-md", s(8, 28, 0.42)},
		{"shadow-lg", s(24, 68, 0.52)},
		{"code-shadow", s(6, 26, 0.4)},

		{"font-display", Fonts[t.Type.Display].Stack},
		{"font-body", Fonts[t.Type.Body].Stack},
		{"font-mono", Fonts[t.Type.Mono].Stack},
		{"display-weight", strconv.Itoa(t.Type.Weight)},
		{"display-tracking", t.Type.Tracking},
		{"display-case", t.Type.Case},
		{"body-tracking", t.Type.BodyTracking},

		{"decor-line", alphaStr(n.Text, decorLineOp)},
		{"decor-dot", alphaStr(n.Text, decorDotOp)},
		{"decor-glow-1", alphaStr(accent, decorGlow1Op)},
		{"decor-glow-2", alphaStr(accent2, decorGlow2Op)},
		{"decor-glow-3", alphaStr(accent3, decorGlow3Op)},
	}

	maxPad := 0
	for _, l := range lines {
		if len(l[0]) > maxPad {
			maxPad = len(l[0])
		}
	}
	var sb strings.Builder
	for i, l := range lines {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("  --%s:%s%s;", l[0], strings.Repeat(" ", maxPad-len(l[0])+1), l[1]))
	}
	return sb.String()
}

func ThemeRootCSS(theme string) string {
	t := Themes[ResolveThemeName(theme)]
	return fmt.Sprintf(":root {\n%s\n}", ThemeVars(t))
}

func ThemeSwitchableCSS() string {
	blocks := make([]string, len(ThemeIDs))
	for i, id := range ThemeIDs {
		selector := fmt.Sprintf(":root[data-theme=\"%s\"]", id)
		if id == DefaultTheme {
			selector = fmt.Sprintf(":root, :root[data-theme=\"%s\"]", id)
		}
		blocks[i] = fmt.Sprintf("%s {\n%s\n}", selector, ThemeVars(Themes[id]))
	}
	return strings.Join(blocks, "\n\n")
}

func HLJSHref(theme string) string {
	return Themes[ResolveThemeName(theme)].HLJS
}

func HLJSMapJSON() string {
	m := make(map[string]string, len(ThemeIDs))
	for _, id := range ThemeIDs {
		m[id] = Themes[id].HLJS
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func DecorMapJSON() string {
	m := make(map[string]string, len(ThemeIDs))
	for _, id := range ThemeIDs {
		m[id] = Themes[id].Decor
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func DecorOf(theme string) string {
	return Themes[ResolveThemeName(theme)].Decor
}

func GoogleFontsHref(themes []string, extra []string) string {
	if len(themes) == 0 {
		themes = ThemeIDs
	}
	wantedMap := make(map[string]bool)
	wantedMap["jetbrains"] = true
	for _, input := range themes {
		t := Themes[ResolveThemeName(input)]
		wantedMap[t.Type.Display] = true
		wantedMap[t.Type.Body] = true
		wantedMap[t.Type.Mono] = true
	}
	for _, key := range extra {
		if key != "" {
			if _, ok := Fonts[key]; ok {
				wantedMap[key] = true
			}
		}
	}
	var wanted []string
	for k := range wantedMap {
		wanted = append(wanted, Fonts[k].Param)
	}
	sort.Strings(wanted)
	for i, p := range wanted {
		wanted[i] = "family=" + p
	}
	return "https://fonts.googleapis.com/css2?" + strings.Join(wanted, "&") + "&display=swap"
}

func FontOverrideCSS() string {
	var blocks []string
	for _, id := range FontIDs {
		blocks = append(blocks, fmt.Sprintf(":root[data-head=\"%s\"] { --font-display: %s; }", id, Fonts[id].Stack))
	}
	for _, id := range FontIDs {
		blocks = append(blocks, fmt.Sprintf(":root[data-body=\"%s\"] { --font-body: %s; }", id, Fonts[id].Stack))
	}
	return "/* ── Font overrides ───────────────────────────────────────────────────── */\n" + strings.Join(blocks, "\n")
}

const DecorCSS = `/* ── Backdrop geometry ────────────────────────────────────────────────── */
:root[data-decor="none"] {
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent);
  --dc-glow-size: 74vmax 74vmax, 58vmax 58vmax;
  --dc-glow-pos: 92% -20%, -14% 112%;
  --dc-glow-anim: dc-breathe 34s ease-in-out infinite;
}

:root[data-decor="grid"] {
  --dc-layers:
    linear-gradient(to right, var(--decor-line) 1px, transparent 1px),
    linear-gradient(to bottom, var(--decor-line) 1px, transparent 1px);
  --dc-size: 72px 72px;
  --dc-mask: radial-gradient(ellipse 88% 76% at 50% 40%, #000 24%, transparent 82%);
  --dc-anim: dc-pan 150s linear infinite;
  --dc-dx: 72px;
  --dc-dy: 72px;
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent);
  --dc-glow-size: 70vmax 70vmax, 56vmax 56vmax;
  --dc-glow-pos: 90% -22%, -12% 110%;
  --dc-glow-anim: dc-breathe 30s ease-in-out infinite;
}

:root[data-decor="dots"] {
  --dc-layers: radial-gradient(var(--decor-dot) 1.4px, transparent 1.6px);
  --dc-size: 30px 30px;
  --dc-mask: radial-gradient(ellipse 92% 80% at 50% 42%, #000 20%, transparent 84%);
  --dc-anim: dc-pan 160s linear infinite;
  --dc-dx: 30px;
  --dc-dy: 30px;
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-3), transparent);
  --dc-glow-size: 64vmax 64vmax, 50vmax 50vmax;
  --dc-glow-pos: -14% -18%, 106% 96%;
  --dc-glow-anim: dc-breathe 36s ease-in-out infinite;
}

:root[data-decor="topo"] {
  --dc-layers: repeating-linear-gradient(52deg,
    var(--decor-line) 0 1px,
    transparent 1px 22px);
  --dc-mask: radial-gradient(ellipse 92% 84% at 46% 44%, #000 18%, transparent 86%);
  --dc-anim: dc-pan 90s linear infinite;
  --dc-dx: 17.34px;
  --dc-dy: -13.54px;
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent);
  --dc-glow-size: 72vmax 72vmax, 52vmax 52vmax;
  --dc-glow-pos: 96% 8%, -10% 104%;
  --dc-glow-anim: dc-breathe 32s ease-in-out infinite;
}

:root[data-decor="beams"] {
  --dc-layers: repeating-linear-gradient(105deg,
    transparent 0 44px,
    var(--decor-glow-1) 44px 118px,
    transparent 118px 166px,
    var(--decor-glow-2) 166px 214px,
    transparent 214px 300px);
  --dc-mask: linear-gradient(to bottom, transparent, #000 22%, #000 70%, transparent);
  --dc-inset: -30%;
  --dc-anim: dc-pan 44s linear infinite;
  --dc-dx: 289.8px;
  --dc-dy: 77.6px;
  --dc-glow: radial-gradient(closest-side, var(--decor-glow-3), transparent);
  --dc-glow-size: 80vmax 80vmax;
  --dc-glow-pos: 50% 118%;
  --dc-glow-anim: dc-breathe 26s ease-in-out infinite;
}

:root[data-decor="rings"] {
  --dc-layers: repeating-radial-gradient(circle at 50% 50%,
    transparent 0 78px,
    var(--decor-line) 78px 79px,
    transparent 79px 158px);
  --dc-size: 100% 100%;
  --dc-repeat: no-repeat;
  --dc-mask: radial-gradient(circle at 50% 46%, #000 12%, transparent 74%);
  --dc-inset: -45%;
  --dc-anim: dc-zoom 52s ease-in-out infinite;
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent);
  --dc-glow-size: 62vmax 62vmax, 48vmax 48vmax;
  --dc-glow-pos: 8% 4%, 96% 92%;
  --dc-glow-anim: dc-breathe 38s ease-in-out infinite;
}

:root[data-decor="waves"] {
  --dc-layers:
    radial-gradient(130% 62% at 50% 128%, transparent 57.6%, var(--decor-line) 58%, transparent 58.4%),
    radial-gradient(130% 66% at 46% 140%, transparent 57.6%, var(--decor-line) 58%, transparent 58.4%),
    radial-gradient(140% 72% at 54% 154%, transparent 57.6%, var(--decor-line) 58%, transparent 58.4%),
    radial-gradient(150% 80% at 50% 172%, transparent 57.6%, var(--decor-line) 58%, transparent 58.4%);
  --dc-size: 100% 100%;
  --dc-repeat: no-repeat;
  --dc-inset: -14%;
  --dc-anim: dc-sway 40s ease-in-out infinite;
  --dc-glow:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-3), transparent);
  --dc-glow-size: 76vmax 60vmax, 54vmax 54vmax;
  --dc-glow-pos: 50% 116%, 88% -16%;
  --dc-glow-anim: dc-breathe 30s ease-in-out infinite;
}

:root[data-decor="orbs"] {
  --dc-layers:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent),
    radial-gradient(closest-side, var(--decor-glow-3), transparent);
  --dc-size: 64vmax 64vmax, 50vmax 50vmax, 42vmax 42vmax;
  --dc-pos: 10% 6%, 86% 28%, 60% 98%;
  --dc-repeat: no-repeat;
  --dc-inset: -22%;
  --dc-anim: dc-float 56s ease-in-out infinite;
}

:root[data-decor="mesh"] {
  --dc-layers:
    radial-gradient(closest-side, var(--decor-glow-1), transparent),
    radial-gradient(closest-side, var(--decor-glow-2), transparent),
    radial-gradient(closest-side, var(--decor-glow-3), transparent),
    radial-gradient(var(--decor-dot) 1px, transparent 1.4px);
  --dc-size: 58vmax 58vmax, 54vmax 54vmax, 46vmax 46vmax, 26px 26px;
  --dc-pos: -8% -14%, 98% 16%, 44% 106%, 0 0;
  --dc-repeat: no-repeat, no-repeat, no-repeat, repeat;
  --dc-inset: -20%;
  --dc-anim: dc-float 62s ease-in-out infinite;
}

:root[data-decor="aurora"] {
  --dc-layers:
    radial-gradient(ellipse 72% 34% at 20% 20%, var(--decor-glow-1), transparent 68%),
    radial-gradient(ellipse 62% 30% at 78% 34%, var(--decor-glow-2), transparent 68%),
    radial-gradient(ellipse 94% 28% at 50% 84%, var(--decor-glow-3), transparent 70%);
  --dc-size: 100% 100%;
  --dc-repeat: no-repeat;
  --dc-inset: -18%;
  --dc-anim: dc-aurora 58s ease-in-out infinite;
}

#backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  overflow: hidden;
  pointer-events: none;
  -webkit-mask-image: var(--dc-mask, none);
          mask-image: var(--dc-mask, none);
}

#backdrop::before,
#backdrop::after {
  content: '';
  position: absolute;
  inset: var(--dc-inset, -12%);
  background-repeat: var(--dc-repeat, repeat);
  will-change: transform;
}

#backdrop::before {
  background-image: var(--dc-layers, none);
  background-size: var(--dc-size, auto);
  background-position: var(--dc-pos, 0 0);
  animation: var(--dc-anim, none);
}

#backdrop::after {
  background-image: var(--dc-glow, none);
  background-size: var(--dc-glow-size, auto);
  background-position: var(--dc-glow-pos, 0 0);
  background-repeat: no-repeat;
  animation: var(--dc-glow-anim, none);
}

/* One tile of travel per cycle, so the pattern lands back on itself. */
@keyframes dc-pan {
  to { transform: translate3d(var(--dc-dx, 0px), var(--dc-dy, 0px), 0); }
}

@keyframes dc-float {
  0%, 100% { transform: translate3d(-1.5%, -1%, 0) scale(1); }
  34%      { transform: translate3d(2%, 1.5%, 0)   scale(1.06); }
  67%      { transform: translate3d(-1%, 2%, 0)    scale(0.97); }
}

@keyframes dc-sway {
  0%, 100% { transform: translate3d(-2.5%, 0, 0) scale(1); }
  50%      { transform: translate3d(2.5%, 1.5%, 0) scale(1.04); }
}

@keyframes dc-zoom {
  0%, 100% { transform: scale(1); }
  50%      { transform: scale(1.1); }
}

@keyframes dc-aurora {
  0%, 100% { transform: translate3d(0, 0, 0) skewY(0deg); }
  50%      { transform: translate3d(3%, -2%, 0) skewY(-1.6deg); }
}

@keyframes dc-breathe {
  0%, 100% { opacity: 0.75; transform: scale(1); }
  50%      { opacity: 1;    transform: scale(1.07); }
}

/* A drifting backdrop is exactly the kind of motion this setting turns off. */
@media (prefers-reduced-motion: reduce) {
  #backdrop::before,
  #backdrop::after { animation: none !important; }
}

@media print {
  /* Fixed elements do not repeat across pages, so the backdrop is rebuilt on
     each slide instead — same custom properties, no second definition. */
  #backdrop { display: none !important; }

  .slide::before,
  .slide::after {
    content: '';
    position: absolute;
    inset: 0;
    z-index: 0;
    animation: none !important;
    -webkit-mask-image: var(--dc-mask, none);
            mask-image: var(--dc-mask, none);
  }

  .slide::before {
    background-image: var(--dc-layers, none);
    background-size: var(--dc-size, auto);
    background-position: var(--dc-pos, 0 0);
    background-repeat: var(--dc-repeat, repeat);
  }

  .slide::after {
    background-image: var(--dc-glow, none);
    background-size: var(--dc-glow-size, auto);
    background-position: var(--dc-glow-pos, 0 0);
    background-repeat: no-repeat;
  }
}`
