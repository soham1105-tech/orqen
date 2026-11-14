package testrun

import "testing"

const realPytestOutput = "..F                                                                      [100%]\n" +
	"=================================== FAILURES ===================================\n" +
	"________________________________ test_sum_wrong ________________________________\n" +
	"\n" +
	"    def test_sum_wrong():\n" +
	"        # Intentional failure so the `orqen test` demo starts at 2 passed / 1 failed.\n" +
	"        # FIX ME: change the expectation to 7 and rerun.\n" +
	"        r = client.get(\"/sum/3/4\")\n" +
	"        assert r.status_code == 200\n" +
	">       assert r.json() == {\"sum\": 8}  # wrong on purpose: /sum/3/4 returns 7\n" +
	"        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^\n" +
	"E       AssertionError: assert {'sum': 7} == {'sum': 8}\n" +
	"E         \n" +
	"E         Differing items:\n" +
	"E         {'sum': 7} != {'sum': 8}\n" +
	"E         Use -v to get more diff\n" +
	"\n" +
	"tests/test_app.py:30: AssertionError\n" +
	"=============================== warnings summary ===============================\n" +
	"../opt/venv/lib/python3.11/site-packages/fastapi/testclient.py:1\n" +
	"  /opt/venv/lib/python3.11/site-packages/fastapi/testclient.py:1: StarletteDeprecationWarning: Using `httpx` with `starlette.testclient` is deprecated; install `httpx2` instead.\n" +
	"    from starlette.testclient import TestClient as TestClient  # noqa\n" +
	"\n" +
	"-- Docs: https://docs.pytest.org/en/stable/how-to/capture-warnings.html\n" +
	"=========================== short test summary info ============================\n" +
	"FAILED tests/test_app.py::test_sum_wrong - AssertionError: assert {'sum': 7} ...\n" +
	"1 failed, 2 passed, 1 warning in 0.49s\n"

func TestParseRealPytestOutput(t *testing.T) {
	passed, failed, skipped := Parse(RunnerPytest, realPytestOutput)
	if passed != 2 || failed != 1 || skipped != 0 {
		t.Errorf("Parse(real pytest) = (%d, %d, %d), want (2, 1, 0)", passed, failed, skipped)
	}
	if got := RunnerFor("python -m pytest -q"); got != RunnerPytest {
		t.Errorf("RunnerFor = %q, want pytest", got)
	}
}
