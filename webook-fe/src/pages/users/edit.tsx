import React, {useEffect, useState} from 'react';
import {Button, DatePicker, Form, Input} from 'antd';
import axios from "@/axios/axios";
import moment from 'moment';
import router from "next/router";

const { TextArea } = Input;

const onFinish = (values: any) => {
    if (values.birthday) {
        values.birthday = moment(values.birthday).format("YYYY-MM-DD")
    }
    axios.post("/users/edit", values)
        .then((res) => {
            if(res.status != 200) {
                alert(res.statusText);
                return
            }
            if (res.data?.code == 0) {
                router.push('/users/profile')
                return
            }
            alert(res.data?.msg || "系统错误");
        }).catch((err) => {
        alert(err);
    })
};

const onFinishFailed = (errorInfo: any) => {
    alert("输入有误")
};

function EditForm() {
    const p: Profile = {} as Profile
    const [data, setData] = useState<Profile>(p)
    const [isLoading, setLoading] = useState(false)

    useEffect(() => {
        setLoading(true)
        axios.get('/users/profile')
            .then((res) => res.data)
            .then((data) => {
                setData(data)
                setLoading(false)
            })
    }, [])

    if (isLoading) return <p>Loading...</p>
    if (!data) return <p>No profile data</p>
    return <Form
        name="basic"
        labelCol={{ span: 8 }}
        wrapperCol={{ span: 16 }}
        style={{ maxWidth: 600 }}
        initialValues={{
            birthday: moment(data.Birthday, 'YYYY-MM-DD'),
            nickname: data.Nickname,
            aboutMe: data.AboutMe
        }}
        onFinish={onFinish}
        onFinishFailed={onFinishFailed}
        autoComplete="off"
    >
        <Form.Item
            label="昵称"
            name="nickname"
        >
            <Input />
        </Form.Item>

        <Form.Item
            label="生日"
            name="birthday"
        >
            <DatePicker format={"YYYY-MM-DD"}
                        placeholder={""}/>
        </Form.Item>

        <Form.Item
            label="关于我"
            name="aboutMe"
        >
            <TextArea rows={4}/>
        </Form.Item>

        <Form.Item wrapperCol={{ offset: 8, span: 16 }}>
            <Button type="primary" htmlType="submit">
                提交
            </Button>
        </Form.Item>
    </Form>
}

export default EditForm;