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
		Name:    "Adobe OAuth Client Secret",
		Pattern: `p8e-[a-z0-9-]{32}`,
	},
	{
		Name:    "Slack Token",
		Pattern: `xox[pbar]-[0-9]{12}-[0-9]{12}-[0-9a-zA-Z]{24}`,
	},
	{
		Name:    "Anthropic API Key",
		Pattern: `sk-ant-api[0-9]{2}-[a-zA-Z0-9_-]{95}`,
	},
	{
		Name:    "Artifcator API Key",
		Pattern: `AKC[a-z0-9]{70}`,
	},
	{
		Name:    "AWS AppSync API Key",
		Pattern: `da2-[a-z0-9]{26}`,
	},
	{
		Name:    "AWS Access Key ID",
		Pattern: `(?:A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}`,
	},
	{
		Name:    "AWS Session Token",
		Pattern: `(?i)(?:aws.?session|aws.?session.?token|aws.?token)(?-i)(?:[" =\\s]*.[" =\\s])[a-zA-Z0-9\\/+=]{100,}(?:[^a-zA-Z\\/=]|\\z)`,
	},
	{
		Name:    "BitBucket Password",
		Pattern: `ATBB[a-zA-Z0-9]{32}`,
	},
	{
		Name:    "DataBricks PAT",
		Pattern: `dapi[a-f0-9]{32}`,
	},
	{
		Name:    "Digital Ocean App Access Token",
		Pattern: `do[opr]_v1_[a-f0-9]{64}`,
	},
	{
		Name:    "Docker PAT",
		Pattern: `dckr_pat_[a-zA-Z0-9_-]{27}`,
	},
	{
		Name:    "DropBox Access Token",
		Pattern: `sl\\.[a-zA-Z0-9_-]{130,152}`,
	},
	{
		Name:    "GitHub App|Refresh|Personal|OAuth Access Token",
		Pattern: `(?:(?:gho|ghp|ghu|ghs)_[a-zA-Z0-9]{36})|(?:ghr_[a-zA-Z0-9]{76})|(?:github_pat_[a-zA-Z0-9_]{82})`,
	},
	// Inferred Matches
	{
		Name:    "AWS Secret Access Key",
		Pattern: `\\baws_?(secret)?_?(access)?_?(key)?.?\\s{0,30}.?\\s{0,30}.?([a-z0-9/+=]{40})\\b`,
	},
	{
		Name:    "Generic Secrets & Keys",
		Pattern: `(?:(?i)password|pass|pw|secret|key|api|access(?-i)).?(?:(?i)key(?-i))?[=",: \\s]{0,20}(?:[0-9a-z]{10,64})`,
	},
}
