import React, {useState, useEffect} from 'react';
import axios, {Result} from "@/axios/axios";
import {Button, Modal, QRCode, Typography} from "antd";
import {ProLayout} from "@ant-design/pro-components";
import {EyeOutlined, LikeOutlined, MoneyCollectOutlined, StarOutlined} from "@ant-design/icons";
import {useSearchParams} from "next/navigation";

export const dynamic = 'force-dynamic'

interface CodeURL {
    codeURL: string
    rid: number
}


function Page(){
    const [data, setData] = useState<Article>()
    const [openQRCode, setOpenQRCode] = useState(false)
    const [codeURL, setCodeURL] = useState('')
    const [isLoading, setLoading] = useState(false)
    const params = useSearchParams()
    // const router = useRouter()
    // const artID = router.query.id
    const artID = params?.get('id') || '1'
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
    let rid = 0
    const reward = function () {
        axios.post<Result<CodeURL>>('/articles/pub/reward', {
            id: parseInt(artID),
            // 固定一分钱
            amt: 1,
        })
            .then((res) => res.data)
            .then((res) => {
                setCodeURL(res.data.codeURL)
                rid = res.data.rid
                setOpenQRCode(true)
            })
    }

    const closeModal = () => {
        setOpenQRCode(false)
        if(rid > 0) {
            axios.post<Result<string>>('/reward/detail', {
                rid: rid,
            }).then((res) => res.data)
                .then((res) => {
                    // 成功了
                    if(res.data == 'RewardStatusPayed') {
                        alert("打赏成功")
                    } else {
                        console.log(res.data)
                    }
                })
        }
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

            <Modal title="扫描二维码" open={openQRCode} onCancel={closeModal} onOk={closeModal}>
                <QRCode value={codeURL} size={128} />
            </Modal>

            <Button icon={<EyeOutlined />}>&nbsp;{data.readCnt}</Button>&nbsp;&nbsp;
            <Button onClick={reward} icon={<MoneyCollectOutlined />}>打赏一分钱</Button>&nbsp;&nbsp;
            <Button onClick={like} icon={<LikeOutlined style={data.liked? {color: "red"}:{}}/>}>&nbsp;{data.likeCnt}</Button>&nbsp;&nbsp;
            <Button onClick={collect} icon={<StarOutlined style={data.collected? {color: "red"}:{}}/>}>&nbsp;{data.collectCnt}</Button>
        </ProLayout>
    )
}

export default Page