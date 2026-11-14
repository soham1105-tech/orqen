from fastapi import FastAPI

app = FastAPI()


@app.get("/")
def root():
    return {"ok": True}


@app.get("/sum/{a}/{b}")
def add(a: int, b: int):
    return {"sum": a + b}
