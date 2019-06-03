#!/usr/bin/env python
import app


if __name__ == "__main__":
    app = app.create_app()
    print("Running app")
    app.run(use_reloader=True, port=5000, host="0.0.0.0", threaded=True)
