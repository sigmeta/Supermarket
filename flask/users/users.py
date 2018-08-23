from flask import Blueprint

users_bp = Blueprint('users', __name__)


@users_bp.route('/')
def hello_world():
    return 'Hello User!'
