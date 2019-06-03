from flask import Flask, jsonify, send_from_directory, request, Response, send_file, flash, redirect
from flask_cors import CORS


def create_app():
    app = Flask(__name__)
    CORS(app)

    @app.route('/', methods=['GET'])
    def hello_world():
        return "Hello world"

    return app
