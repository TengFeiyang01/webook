'use client';
import {EditOutlined} from '@ant-design/icons';
import {ProLayout, ProList} from '@ant-design/pro-components';
import {Button, Tag} from 'antd';
import React, {useEffect, useState} from 'react';
import axios from "@/axios/axios";
import router from "next/router";

const IconText = ({ icon, text, onClick }: { icon: any; text: string, onClick: any}) => (
    <Button onClick={onClick} type={"default"}>
    {React.createElement(icon, { style: { marginInlineEnd: 8 } })}
        {text}
  </Button>
);

interface ArticleItem {
    id: bigint
    title: string
    status: number
    abstract: string
}

const ArticleList = () => {
    const [data, setData] = useState<Array<ArticleItem>>([])
    const [loading, setLoading] = useState<boolean>()
    useEffect(() => {
        setLoading(true)
        axios.post('/articles/list', {
            "offset": 0,
            "limit": 100,
        }).then((res) => res.data)
            .then((data) => {
                setData(data.data)
                setLoading(false)
            })
    }, [])
    return (
        <ProLayout title={"创作中心"}>
            <ProList<ArticleItem>
                toolBarRender={() => {
                    return [
                        <Button key="3" type="primary" href={"/articles/edit"}>
                            写作
                        </Button>,
                    ];
                }}
                itemLayout="vertical"
                rowKey="id"
                headerTitle="文章列表"
                loading={loading}
                dataSource={data}
                // ts:ignore
                metas={{
                    title: {
                        dataIndex: "title"
                    },
                    description: {
                        render: (data, record, idx) => {
                            switch (record.status) {
                                case 1:
                                    return (
                                        <Tag color={"processing"}>未发表</Tag>
                                    )
                                case 2:
                                    return (
                                        <Tag color={"success"}  >已发表</Tag>
                                    )
                                case 3:
                                    return (
                                        <Tag color={"warning"}>尽自己可见</Tag>
                                    )
                                default:
                                    return (<></>)
                            }

                            },
                    },
                    actions: {
                        render: (text, row) => [
                            <IconText
                                icon={EditOutlined}
                                text="编辑"
                                onClick={() => {
                                    router.push("/articles/edit?id=" + row.id.toString())
                                }}
                                key="list-vertical-edit-o"
                            />,
                        ],
                    },
                    extra: {
                        render: () => (
                            <img
                                width={272}
                                alt="logo"
                                src="https://gw.alipayobjects.com/zos/rmsportal/mqaQswcyDLcXyDKnZfES.png"
                            />
                        ),
                    },
                    content: {
                        render: (node, record) => {
                            return (
                                <div dangerouslySetInnerHTML={{__html: record.abstract}}>
                                </div>
                            )
                        }
                    },
                }}
            />
        </ProLayout>
    );
};

export default ArticleList;
