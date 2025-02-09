import type { Metadata } from 'next'
import {BrowserRouter, Route, Routes} from "react-router-dom";
import React from "react";
export const metadata: Metadata = {
    title: '小微书',
    description: '你的第一个 Web 应用',
}

const App = () => {
    return <div>
       hello
    </div>
}

export default App
