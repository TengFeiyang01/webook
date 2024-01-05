import React, {useState, useEffect, CSSProperties} from 'react';
import axios from "@/axios/axios";
import {useSearchParams} from "next/navigation";
import {Button, Typography} from "antd";
import {ProLayout} from "@ant-design/pro-components";
import {EyeOutlined, LikeOutlined, StarOutlined} from "@ant-design/icons";
import color from "@wangeditor/basic-modules/dist/basic-modules/src/modules/color";

function Page(){
    const [data, setData] = useState<Article>()
    const [isLoading, setLoading] = useState(false)
    const params = useSearchParams()
    const artID = params?.get("id")!
    debugger
    useEffect(() => {
        setLoading(true)
        axios.get('/articles/pub/'+artID)
            .then((res) => res.data)
            .then((data) => {
                setData(data.data)
                setLoading(false)
            })
    }, [artID])

    if (isLoading) return <p>Loading...</p>
    if (!data) return <p>No data</p>

    const like = () => {
        axios.post('/articles/pub/like', {
            id: parseInt(artID),
            like: !data.liked
        })
            .then((res) => res.data)
            .then((res) => {
                if(res.code == 0) {
                    if (data.liked) {
                        data.likeCnt --
                    } else {
                        data.likeCnt ++
                    }
                    data.liked = !data.liked
                    setData(Object.assign({}, data))
                }
            })
    }

    const collect = () => {
        if (data.collected) {
            return
        }
        axios.post('/articles/pub/collect', {
            id: parseInt(artID),
            // 你可以加上增删改查收藏夹的功能，在这里传入收藏夹 ID
            cid: 0,
        })
            .then((res) => res.data)
            .then((res) => {
                if(res.code == 0) {
                    data.collectCnt ++
                    data.collected = !data.collected
                    setData(Object.assign({}, data))
                }
            })
    }

    return (
        <ProLayout pure={true}>
            <Typography>
                <Typography.Title>
                    {data.title}
                </Typography.Title>
                <Typography.Paragraph>
                    <div dangerouslySetInnerHTML={{__html: data.content}}></div>
                </Typography.Paragraph>
            </Typography>
            <Button icon={<EyeOutlined />}>&nbsp;{data.readCnt}</Button>&nbsp;&nbsp;
            <Button onClick={like} icon={<LikeOutlined style={data.liked? {color: "red"}:{}}/>}>&nbsp;{data.likeCnt}</Button>&nbsp;&nbsp;
            <Button onClick={collect} icon={<StarOutlined style={data.collected? {color: "red"}:{}}/>}>&nbsp;{data.collectCnt}</Button>
        </ProLayout>
    )
}

export default Page