import React from 'react';
import './App.css'
import { MatchRoom, QueueTiles } from './ChessUI.tsx'
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";

function App() {
  console.log(import.meta.env.VITE_API_URL)
  console.log(`React Version: ${React.version}`)

  return (
    <Router>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/matchroom/:matchid" element={<MatchRoom />} />
      </Routes>
    </Router>
  )
}

function Home() {

  return (
    <>
      <QueueTiles />
    </>
  )
}

export default App
