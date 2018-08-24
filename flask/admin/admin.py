from flask import (Blueprint, flash, g, redirect, render_template, request, session, url_for, make_response, get_flashed_messages)
import os
import json
import requests


admin_bp = Blueprint('admin', __name__, template_folder='templates',
                     static_folder='static')


#主页
@admin_bp.route('/')
def index():
    if request.cookies.get("name"):
        return render_template("index.html")
    else:
        return render_template("login.html")


#登录页面
@admin_bp.route('/login', methods=['POST', 'GET'])
def login():
    print('login')
    if request.method == 'GET':
        return redirect("admin")
    else:
        #print(request.form)
        user = request.form['username']
        pwd = request.form['password']
        org = request.form['orgName']
        res=requests.post("http://localhost:4000/users",
                          "username=%s&orgName=%s&password=%s"%(user,org,pwd),
                         headers={"content-type": "application/x-www-form-urlencoded"})
        if res.status_code != 200:
            return render_template("error.html",message="status_code: "+res.status_code+res.text)
        restext=json.loads(res.text)
        print(restext)
        if restext['success']!=True:
            return render_template("error.html",message=restext['message'])

        resp = make_response(redirect("admin"))
        resp.set_cookie('name',user)
        resp.set_cookie('token', restext['token'])
        return resp


#注册页
@admin_bp.route('/signup')
def signup():
    return render_template("sign-up.html")


#处理注册
@admin_bp.route('/register', methods=['POST'])
def register():
    user = request.form['username']
    pwd = request.form['password']
    org = request.form['orgName']
    res = requests.post("http://localhost:4000/register",
                        "username=%s&orgName=%s&password=%s" % (user, org, pwd),
                        headers={"content-type": "application/x-www-form-urlencoded"})
    if res.status_code != 200:
        return render_template("error.html", message="status_code: " + res.status_code + res.text)
    restext = json.loads(res.text)
    print(restext)
    if restext['success'] != True:
        return render_template("error.html", message=restext['message'])

    resp = make_response(redirect("admin"))
    resp.set_cookie('name', user)
    resp.set_cookie('token', restext['token'])
    return resp


#处理退出
@admin_bp.route('/logout', methods=['GET'])
def logout():
    resp = make_response(redirect("admin"))
    resp.delete_cookie('name')
    resp.delete_cookie('token')
    return resp


#错误页面
@admin_bp.route('/error/<message>')
def error(message):
    return render_template("error.html",message=message)


#查货表单
