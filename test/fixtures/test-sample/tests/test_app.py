"""Demo test suite for `orqen test`.

Two tests pass and one fails on purpose: run `orqen test` to see the failure,
fix `test_sum_wrong` below, then run it again to see all three pass.
"""
from fastapi.testclient import TestClient

from main import app

client = TestClient(app)


def test_root_ok():
    r = client.get("/")
    assert r.status_code == 200
    assert r.json() == {"ok": True}


def test_sum_ok():
    r = client.get("/sum/2/3")
    assert r.status_code == 200
    assert r.json() == {"sum": 5}


def test_sum_wrong():
    # Intentional failure so the `orqen test` demo starts at 2 passed / 1 failed.
    # FIX ME: change the expectation to 7 and rerun.
    r = client.get("/sum/3/4")
    assert r.status_code == 200
    assert r.json() == {"sum": 8}  # wrong on purpose: /sum/3/4 returns 7
