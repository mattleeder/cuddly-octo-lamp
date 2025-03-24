import React, { useState } from "react"

enum Page {
  All = "All",
  Bullet = "Bullet",
  Blitz = "Blitz",
  Rapid = "Rapid",
  Classical = "Classical",
}

function AccountSidebar({ activePage, setActivePage }: { activePage: Page, setActivePage: React.Dispatch<React.SetStateAction<Page>> }) {
  return (
    <div style={{
      display: "block",
      position: "fixed",
      left: 0,
      top: "3em",
      textAlign: "left",
      padding: 0,
      margin: 0,
    }}>
      <ul style={{
        display: "block",
        listStyle: "none",
        padding: 0,
        margin: 0,
      }}>
        <li className={`userTimeFormat${activePage == Page.All ? " active" : ""}`} onClick={() => setActivePage(Page.All)}>
          All
        </li>
        <li className={`userTimeFormat${activePage == Page.Bullet ? " active" : ""}`} onClick={() => setActivePage(Page.Bullet)}>
          Bullet
        </li>
        <li className={`userTimeFormat${activePage == Page.Blitz ? " active" : ""}`} onClick={() => setActivePage(Page.Blitz)}>
          Blitz
        </li>
        <li className={`userTimeFormat${activePage == Page.Rapid ? " active" : ""}`} onClick={() => setActivePage(Page.Rapid)}>
          Rapid
        </li>
        <li className={`userTimeFormat${activePage == Page.Classical ? " active" : ""}`} onClick={() => setActivePage(Page.Classical)}>
          Classical
        </li>
      </ul>
    </div>
  )
}

function AccountContent({ page }: { page: Page }) {
  return (
    <div>
      {page}
    </div>
  )
}

export function AccountPage() {
  const [activePage, setActivePage] = useState(Page.All)
    return (
      <div>
      <AccountSidebar activePage={activePage} setActivePage={setActivePage}/>
      <AccountContent page={activePage}/>
      </div>
    )
}