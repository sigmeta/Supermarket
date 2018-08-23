from flask import (Blueprint, flash, g, redirect, render_template, request, session, url_for, make_response)
import os
import json


admin_bp = Blueprint('admin', __name__, template_folder='templates',
                     static_folder='static')


@admin_bp.route('/')
def index():
    if request.cookies.get("name"):
        return render_template("index.html")
    else:
        return render_template("login.html")


@admin_bp.route('/login', methods=['POST', 'GET'])
def login():
    if request.method == 'GET':
        return redirect("admin")
    else:
        if not os.path.exists("admin.json"):
            with open("admin.json",'w') as f:
                f.write(json.dumps({}))
            return "用户名或密码错误"

        with open("admin.json") as f:
            user_pwd=json.loads(f.read())

        user = request.form['name']
        pwd = request.form['password']
        if user not in user_pwd or user_pwd[user]!=pwd:
            return "用户名或密码错误"

        resp = make_response(redirect("admin"))
        resp.set_cookie('name',user)
        return resp


@admin_bp.route('/signup')
def signup():
    return render_template("sign-up.html")


@admin_bp.route('/register', methods=['POST'])
def register():
    user = request.form['name']
    pwd = request.form['password']
    if not os.path.exists("admin.json"):
        with open("admin.json", 'w') as f:
            f.write(json.dumps({user:pwd}))
        resp = make_response(redirect("admin"))
        resp.set_cookie('name', user)
        return resp
    else:
        with open("admin.json", 'r') as f:
            user_pwd=json.loads(f.read())
        if user in user_pwd:
            return "用户名已存在"

        user_pwd[user]=pwd
        with open("admin.json", 'w') as f:
            f.write(json.dumps(user_pwd))
        resp = make_response(redirect("admin"))
        resp.set_cookie('name', user)
        return resp


@admin_bp.route('/logout', methods=['GET'])
def logout():
    resp = make_response(redirect("admin"))
    resp.delete_cookie('name')
    return resp


