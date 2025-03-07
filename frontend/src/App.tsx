import React from 'react';
import './App.css'
import { MatchRoom } from './chess/MatchRoom.tsx'
import { QueueTiles } from './chess/QueueTiles.tsx';
import {
  BrowserRouter as Router,
  Routes,
  Route,
} from "react-router-dom";
import { HomepageMatch } from './chess/HomepageMatch.tsx';

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
      <HomepageMatch />
    </>
  )
}

export default App
