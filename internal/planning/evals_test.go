package planning

import "testing"

func TestLoadBenchmarkFixtures(t *testing.T) {
	fixtures, err := LoadBenchmarkFixtures()
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) < 3 {
		t.Fatalf("expected at least 3 benchmark fixtures, got %d", len(fixtures))
	}
}

func TestRubricEvaluationIsDeterministic(t *testing.T) {
	fixtures, err := LoadBenchmarkFixtures()
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		first := EvaluateBenchmarkFixture(fixture)
		second := EvaluateBenchmarkFixture(fixture)
		if first.Scores != second.Scores {
			t.Fatalf("expected deterministic scores for %s: %+v != %+v", fixture.Slug, first.Scores, second.Scores)
		}
	}
}

func TestBenchmarkFixturesSatisfyMinimumScores(t *testing.T) {
	fixtures, err := LoadBenchmarkFixtures()
	if err != nil {
		t.Fatal(err)
	}
	for _, fixture := range fixtures {
		evaluation := EvaluateBenchmarkFixture(fixture)
		if !evaluation.Scores.MeetsMinimum(fixture.MinimumScores) {
			t.Fatalf("fixture %s failed minimum scores: got %+v want %+v", fixture.Slug, evaluation.Scores, fixture.MinimumScores)
		}
	}
}
