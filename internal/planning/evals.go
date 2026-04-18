package planning

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"plan/internal/notes"
)

type RubricScores struct {
	Clarity             int `json:"clarity"`
	Boundedness         int `json:"boundedness"`
	RiskCoverage        int `json:"risk_coverage"`
	NonGoalQuality      int `json:"non_goal_quality"`
	VerificationQuality int `json:"verification_quality"`
	StorySliceQuality   int `json:"story_slice_quality"`
	AgentExecutability  int `json:"agent_executability"`
	ArtifactSimplicity  int `json:"artifact_simplicity"`
	UserEffortCost      int `json:"user_effort_cost"`
}

type BenchmarkFixture struct {
	Slug          string       `json:"slug"`
	Title         string       `json:"title"`
	ArtifactType  string       `json:"artifact_type"`
	Scenario      string       `json:"scenario"`
	Candidate     string       `json:"candidate"`
	MinimumScores RubricScores `json:"minimum_scores"`
}

type BenchmarkEvaluation struct {
	Fixture BenchmarkFixture `json:"fixture"`
	Scores  RubricScores     `json:"scores"`
}

func (s RubricScores) MeetsMinimum(min RubricScores) bool {
	return s.Clarity >= min.Clarity &&
		s.Boundedness >= min.Boundedness &&
		s.RiskCoverage >= min.RiskCoverage &&
		s.NonGoalQuality >= min.NonGoalQuality &&
		s.VerificationQuality >= min.VerificationQuality &&
		s.StorySliceQuality >= min.StorySliceQuality &&
		s.AgentExecutability >= min.AgentExecutability &&
		s.ArtifactSimplicity >= min.ArtifactSimplicity &&
		s.UserEffortCost >= min.UserEffortCost
}

func LoadBenchmarkFixtures() ([]BenchmarkFixture, error) {
	dir, err := benchmarkFixtureDir()
	if err != nil {
		return nil, err
	}
	return LoadBenchmarkFixturesFromDir(dir)
}

func LoadBenchmarkFixturesFromDir(dir string) ([]BenchmarkFixture, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read benchmark fixtures: %w", err)
	}
	var fixtures []BenchmarkFixture
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read benchmark fixture %s: %w", entry.Name(), err)
		}
		var fixture BenchmarkFixture
		if err := json.Unmarshal(raw, &fixture); err != nil {
			return nil, fmt.Errorf("parse benchmark fixture %s: %w", entry.Name(), err)
		}
		if strings.TrimSpace(fixture.Slug) == "" || strings.TrimSpace(fixture.Candidate) == "" {
			return nil, fmt.Errorf("benchmark fixture %s is missing required content", entry.Name())
		}
		fixtures = append(fixtures, fixture)
	}
	sort.Slice(fixtures, func(i, j int) bool {
		return fixtures[i].Slug < fixtures[j].Slug
	})
	return fixtures, nil
}

func EvaluateBenchmarkFixture(fixture BenchmarkFixture) BenchmarkEvaluation {
	candidate := fixture.Candidate
	clarity := scoreClarity(candidate)
	boundedness := scoreBoundedness(candidate)
	riskCoverage := scoreRiskCoverage(candidate)
	nonGoalQuality := scoreNonGoalQuality(candidate)
	verificationQuality := scoreVerificationQuality(candidate)
	storySliceQuality := scoreStorySliceQuality(candidate)
	agentExecutability := clampScore((clarity+boundedness+verificationQuality)/3 + boolScore(hasStrongSection(candidate, "Solution Shape") || hasStrongSection(candidate, "Description")), 0, 5)
	artifactSimplicity := scoreArtifactSimplicity(candidate)
	userEffortCost := scoreUserEffortCost(candidate)

	return BenchmarkEvaluation{
		Fixture: fixture,
		Scores: RubricScores{
			Clarity:             clarity,
			Boundedness:         boundedness,
			RiskCoverage:        riskCoverage,
			NonGoalQuality:      nonGoalQuality,
			VerificationQuality: verificationQuality,
			StorySliceQuality:   storySliceQuality,
			AgentExecutability:  agentExecutability,
			ArtifactSimplicity:  artifactSimplicity,
			UserEffortCost:      userEffortCost,
		},
	}
}

func benchmarkFixtureDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("resolve benchmark fixture dir")
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(file)))
	return filepath.Join(root, "testdata", "evals", "fixtures"), nil
}

func scoreClarity(candidate string) int {
	score := 0
	if hasStrongSection(candidate, "Problem") {
		score += 2
	}
	if hasStrongSection(candidate, "Goals") {
		score += 2
	}
	if hasStrongSection(candidate, "Why") || hasStrongSection(candidate, "Outcome") {
		score++
	}
	return clampScore(score, 0, 5)
}

func scoreBoundedness(candidate string) int {
	score := 0
	if hasStrongSection(candidate, "Constraints") {
		score++
	}
	if hasStrongSection(candidate, "Non-Goals") {
		score += 2
	}
	if hasStrongSection(candidate, "Scope Boundary") || hasStrongSection(candidate, "Appetite") {
		score += 2
	}
	return clampScore(score, 0, 5)
}

func scoreRiskCoverage(candidate string) int {
	score := 0
	risks := notes.ExtractSection(candidate, "Risks / Open Questions")
	if strings.TrimSpace(risks) != "" {
		score += 2
	}
	if !sectionLooksThin(risks) {
		score++
	}
	if containsAny(strings.ToLower(candidate), []string{"risk", "rollback", "failure", "assumption", "rabbit hole", "dependency"}) {
		score += 2
	}
	return clampScore(score, 0, 5)
}

func scoreNonGoalQuality(candidate string) int {
	nonGoals := notes.ExtractSection(candidate, "Non-Goals")
	score := 0
	if strings.TrimSpace(nonGoals) != "" {
		score += 2
	}
	if !sectionLooksThin(nonGoals) {
		score += 2
	}
	if containsAny(strings.ToLower(nonGoals), []string{"not", "exclude", "avoid", "no "}) {
		score++
	}
	return clampScore(score, 0, 5)
}

func scoreVerificationQuality(candidate string) int {
	verification := notes.ExtractSection(candidate, "Verification")
	score := 0
	if strings.TrimSpace(verification) != "" {
		score += 2
	}
	if !sectionLooksThin(verification) {
		score++
	}
	if containsAny(strings.ToLower(verification), []string{"test", "validate", "manual", "verify", "assert", "check"}) {
		score += 2
	}
	return clampScore(score, 0, 5)
}

func scoreStorySliceQuality(candidate string) int {
	score := 0
	if hasStrongSection(candidate, "Story Breakdown") || hasStrongSection(candidate, "Acceptance Criteria") {
		score += 2
	}
	if bulletCount(candidate) >= 3 {
		score++
	}
	if containsAny(strings.ToLower(candidate), []string{"story", "slice", "increment", "acceptance"}) {
		score += 2
	}
	return clampScore(score, 0, 5)
}

func scoreArtifactSimplicity(candidate string) int {
	score := 5
	if headingCount(candidate) > 18 {
		score--
	}
	if len(strings.Fields(candidate)) > 350 {
		score--
	}
	if bulletCount(candidate) > 25 {
		score--
	}
	return clampScore(score, 1, 5)
}

func scoreUserEffortCost(candidate string) int {
	score := 5
	if headingCount(candidate) > 16 {
		score--
	}
	if len(strings.Fields(candidate)) > 300 {
		score--
	}
	if bulletCount(candidate) > 20 {
		score--
	}
	return clampScore(score, 1, 5)
}

func hasStrongSection(candidate, heading string) bool {
	section := notes.ExtractSection(candidate, heading)
	if strings.TrimSpace(section) == "" {
		return false
	}
	return !sectionLooksThin(section)
}

func headingCount(candidate string) int {
	count := 0
	for _, line := range strings.Split(candidate, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			count++
		}
	}
	return count
}

func bulletCount(candidate string) int {
	count := 0
	for _, line := range strings.Split(candidate, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "1. ") {
			count++
		}
	}
	return count
}

func clampScore(value, min, max int) int {
	switch {
	case value < min:
		return min
	case value > max:
		return max
	default:
		return value
	}
}

func boolScore(ok bool) int {
	if ok {
		return 1
	}
	return 0
}
