import React, { useState } from "react"

interface SideBarTab {
  title: string,
  content: React.ReactElement,
}

export function SideBar({ tabs }: { tabs: SideBarTab[] }) {
  const [selectedTitle, setSelectedTitle] = useState(tabs[0].title)
  const selectedTab = tabs.find(tab => tab.title === selectedTitle)

  if (tabs.length == 0) {
    throw new Error("Must give at least 1 tab to sidebar.")
  }
  
  return (
    <div className="sidebarContainer">
      <div className="sidebarButtonContainer">      
        {tabs.map((tab) => {
          return (
            <button
              key={tab.title}
              onClick={() => setSelectedTitle(tab.title)}
              className={`sidebarButton${tab.title === selectedTab?.title ? " active": ""}`}
            >
              {tab.title}
            </button>
          )
        })}
      </div>
      <div className="sidebarContent">
        {selectedTab?.content}
      </div>
    </div>
  )
}