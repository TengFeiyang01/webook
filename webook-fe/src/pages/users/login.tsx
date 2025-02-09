import React from 'react';
import { Button, Form, Input } from 'antd';
import axios from "@/axios/axios";
import Link from "next/link";
import router from "next/router";

const onFinish = (values: any) => {
    axios.post("/users/login", values)
        .then((res) => {
            if(res.status != 200) {
                alert(res.statusText);
                return
            }
            if(typeof res.data == 'string') {
                alert(res.data);
            } else {
                const msg = res.data?.msg || JSON.stringify(res.data)
                alert(msg);
                if(res.data.code == 0) {
                    router.push('/articles/list')
                }
            }
        }).catch((err) => {
            alert(err);
    })
};

const onFinishFailed = (errorInfo: any) => {
    alert("输入有误")
};

const LoginForm: React.FC = () => {
    return (<Form
        name="basic"
        labelCol={{ span: 8 }}
        wrapperCol={{ span: 16 }}
        style={{ maxWidth: 600 }}
        initialValues={{ remember: true }}
        onFinish={onFinish}
        onFinishFailed={onFinishFailed}
        autoComplete="off"
    >
        <Form.Item
            label="邮箱"
            name="email"
            rules={[{ required: true, message: '请输入邮箱' }]}
        >
            <Input />
        </Form.Item>

        <Form.Item
            label="密码"
            name="password"
            rules={[{ required: true, message: '请输入密码' }]}
        >
            <Input.Password />
        </Form.Item>

        <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
            <Button type="primary" htmlType="submit">
                登录
            </Button>
            <Link href={"/users/login_sms"} >
                &nbsp;&nbsp;手机号登录
            </Link>
            <Link href={"/users/login_wechat"} >
                &nbsp;&nbsp;微信扫码登录
            </Link>
            <Link href={"/users/signup"} >
                &nbsp;&nbsp;注册
            </Link>
        </Form.Item>
    </Form>
)};

export default LoginForm;