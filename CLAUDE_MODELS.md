# Claude Model IDs (Current as of Dec 2024)

## Latest Models (Claude 4.5 Series)

### Claude Sonnet 4.5
- **API ID**: `claude-sonnet-4-5-20250929`
- **Alias**: `claude-sonnet-4-5`
- **Description**: Smart model for complex agents and coding
- **Pricing**: $3/MTok input, $15/MTok output

### Claude Haiku 4.5
- **API ID**: `claude-haiku-4-5-20251001`
- **Alias**: `claude-haiku-4-5`
- **Description**: Fastest model with near-frontier intelligence
- **Pricing**: $1/MTok input, $5/MTok output

### Claude Opus 4.5
- **API ID**: `claude-opus-4-5-20251101`
- **Alias**: `claude-opus-4-5`
- **Description**: Premium model with maximum intelligence
- **Pricing**: $5/MTok input, $25/MTok output

## Legacy Models (Claude 3.5)

The models we were trying to use are outdated:
- ❌ `claude-3-sonnet-20240229` - No longer exists
- ❌ `claude-3-5-sonnet-20240620` - Deprecated

## Recommendation

For echo9llama, we should use:
- **Primary**: `claude-sonnet-4-5-20250929` (best for coding and agents)
- **Fallback**: `claude-haiku-4-5-20251001` (faster, cheaper)
