package anbuCrypto

const (
	maxSize = 50 * 1024 * 1024 // 50MB
)

// SecretRules defines the patterns to scan for
var secretRules = []struct {
	Name    string
	Pattern string
}{
	// Direct Matches
	{
		Name:    "Slack Token", // Slack Bot|User|Workspace Access|Workspace Refresh Token
		Pattern: `xox[pebar]-[\d]{10,12}-[\d]{10,12}-[\w]{23,32}`,
	},
	{
		Name:    "Slack Webhook URL",
		Pattern: `https://hooks.slack.com/(?:services|workflows|triggers)/[\w+/]{43,56}`,
	},
	{
		Name:    "Anthropic API Key",
		Pattern: `sk-ant-(?:admin|api)[\d]{2}-[\w\-]{93}AA`,
	},
	{
		Name:    "Artifcatory API Key",
		Pattern: `AKC[\w]{70}`,
	},
	{
		Name:    "AWS Access Key ID",
		Pattern: `(?:A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`,
	},
	{
		Name:    "BitBucket Password",
		Pattern: `ATBB[\w]{32}`,
	},
	{
		Name:    "Digital Ocean Access Token",
		Pattern: `(?:do[opr]_v1_[a-f0-9]{64})|(?:do_[a-z]{2,3}_[a-z0-9]{64})`,
	},
	{
		Name:    "Docker PAT",
		Pattern: `dckr_pat_[\w_-]{27}`,
	},
	{
		Name:    "GitHub Token", // GitHub App|Refresh|Personal|OAuth Access Token
		Pattern: `(?:(?:gho|ghp|ghu|ghs)_[\w]{36})|(?:ghr_[\w]{76})|(?:github_pat_[\w_]{82})`,
	},
	{
		Name:    "GitLab Token", // GitLab PAT|Pipeline|Runner Auth Token
		Pattern: `(?:(?:glpat|glrt)-[\w-]{20})|(?:glptt-[0-9a-f]{40})`,
	},
	{
		Name:    "Google API Key",
		Pattern: `AIza[\w-]{35}`,
	},
	{
		Name:    "Google OAuth Access Token",
		Pattern: `ya29\.[\w_-]{20,256}[\w_-]{20,256}[\w_-]{20,256}`, // go has length limit
	},
	{
		Name:    "JSON Web Token (JWT)",
		Pattern: `ey[\w]{17,}\.ey[\w/_-]{17,}\.(?:[\w/_-]{10,}={0,2})?`,
	},
	{
		Name:    "Heroku API Key",
		Pattern: `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
	},
	{
		Name:    "Hugging Face API Token",
		Pattern: `hf_[a-zA-Z]{34}`,
	},
	{
		Name:    "New Relic API Key",
		Pattern: `(?:NRAK|NRAA)-[a-z0-9]{27}`,
	},
	{
		Name:    "Okta Access Token",
		Pattern: `00[\w=-]{40}`,
	},
	{
		Name:    "OpenAI API Key",
		Pattern: `(?:sk-(?:proj|svcacct|admin)-(?:[\w_-]{74}|[\w_-]{58})T3BlbkFJ(?:[\w_-]{74}|[\w_-]{58}))|(?:sk-[\w]{20}T3BlbkFJ[\w]{20})`,
	},
	{
		Name:    "Shopify Token", // Shopify Shared Secret|Access Token|Custom Access Token|Private App Access Token
		Pattern: `(?:(?:shpss|shpat|shpca|shppa)_[a-fA-F0-9]{32})`,
	},
	{
		Name:    "Stripe API Key",
		Pattern: `(?:sk|rk)_(?:test|live|prod)_[\w]{10,99}`,
	},
	// done
	{
		Name:    "Telegram Bot Token",
		Pattern: `[0-9]{5,16}:(?-i:A)[a-z0-9_\\-]{34}`,
	},
	{
		Name:    "Twilio API Key",
		Pattern: `SK[a-f0-9]{32}`,
	},
	{
		Name:    "Twitter Access Token",
		Pattern: `[0-9]{15,25}-[\w]{20,40}`,
	},
	{
		Name:    "Vault Token",
		Pattern: `(?:hvs\.[\w-]{90,120}|s\.(?i:[a-z0-9]{24}))`,
	},
	// Inferred Matches
	{
		Name:    "Cloudflare Global API Key",
		Pattern: `(?i)(?:(?:cloudflare)|(?:cf)).?(?:api)?(?-i).{0,25}\b(?:[a-z0-9]{37})\b`,
	},
	{
		Name:    "Cloudflare API Key",
		Pattern: `(?i)(?:(?:cloudflare)|(?:cf)).?(?:api)?(?-i).{0,25}\b(?:[\w-]{40})\b`,
	},
	{
		Name:    "AWS Secret Access Key",
		Pattern: `(?:)(?i)(?:aws).?(?:secret)?.?(?:access)?.?(?:key)?(?-i).{0,25}\b(?:[a-z0-9/+=]{40})\b`,
	},
	{
		Name:    "AWS Session Token",
		Pattern: `(?i)(?:aws).?(?:session).?(?:token)?(?-i).{0,25}\b(?:[\w/_\.=+-]{100,})\b`,
	},
	{
		Name:    "Generic Secrets & Keys",
		Pattern: `(?:(?i)password|pass|pw|secret|key|api|access(?-i)).?(?:(?i)key(?-i))?.{0,25}\b(?:[\w]{20,64})\b`,
	},
}
