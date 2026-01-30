package config

// KnownMevBots 已知的 MEV Bot 和 Builder 地址
// 数据来源：社区收集和公开数据整理
var KnownMevBots = map[string]string{
	// === MEV Builders (区块构建者) ===
	"0xdafea492d9c6733ae3d56b7ed1adb60692c98bc5": "Flashbots Builder",
	"0xb64a30399f7f6b0c154c2e7af0a3ec7b0a5b131a": "Flashbots Builder (Old)",
	"0xf2f5c73fa04406b1995e397b55c24ab1f3ea726c": "bloXroute Max-Profit",
	"0xf573d99385c05c23b24ed33de616ad16a43a0919": "bloXroute Non-Sandwich",
	"0x199d5ed7f45f4ee35960cf22eade2076e95b253f": "bloXroute Regulated",
	"0xaab27b150451726ec7738aa1d0a94505c8729bd1": "Eden Network",
	"0x95222290dd7278aa3ddd389cc1e1d165cc4bafe5": "beaverbuild.org",
	"0x1f9090aae28b8a3dceadf281b0f12828e676c326": "rsync-builder.xyz",
	"0x4838b106fce9647bdf1e7877bf73ce8b0bad5f97": "Titan Builder",

	// === 知名 MEV Searchers (搜索者/Bot) ===
	// jaredfromsubway.eth - 最活跃的三明治攻击 Bot
	"0xae2fc483527b8ef99eb5d9b44875f005ba1fae13": "jaredfromsubway.eth",
	"0x6b75d8af000000e20b7a7ddf000ba900b4009a80": "jaredfromsubway.eth Bot",
	"0x1f2f10d1c40777ae1da742455c65828ff36df387": "jaredfromsubway.eth Bot 2",

	// 其他知名 MEV Bots
	"0xa69babef1ca67a37ffaf7a485dfff3382056e78c": "MEV Bot",
	"0x00000000000007736e2f9af06b8f5f3b6d0e8f13": "MEV Bot",
	"0x000000000000084e91743124a982076c59f10084": "Sandwich Bot",
	"0x00000000003b3cc22af3ae1eac0440bcee416b40": "MEV Bot",
	"0x51c72848c68a965f66fa7a88855f9f7784502a7f": "MEV Searcher",
	"0xd2269f890854a8c5f03e8ea091e3d5a2e0e0f890": "MEV Bot",
}

// MevBotAddressPatterns MEV Bot 地址的常见模式
// 很多 MEV Bot 使用特殊的地址模式（如多个前导零）
var MevBotAddressPatterns = []string{
	"0x000000000000", // 12个前导零
	"0x00000000",     // 8个前导零
}

// IsMevBot 检查地址是否为已知的 MEV Bot
func IsMevBot(address string) (bool, string) {
	if name, exists := KnownMevBots[address]; exists {
		return true, name
	}

	// 检查地址模式
	for _, pattern := range MevBotAddressPatterns {
		if len(address) >= len(pattern) && address[:len(pattern)] == pattern {
			return true, "Potential MEV Bot (Pattern Match)"
		}
	}

	return false, ""
}
