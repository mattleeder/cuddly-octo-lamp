import { Search } from 'lucide-react';
import React, { useEffect, useRef, useState } from 'react';

function NavbarSearchInput({ active, ref }: { active: boolean, ref: React.RefObject<HTMLInputElement | null> }) {
  return (
    <input 
      className="navbarSearchInput"
      style={{width: `${active ? "100%" : "0%"}`}}
      ref={ref}
      placeholder='Search'
    ></input>
  )
}

export function NavbarSearch() {
  const [searchActive, setSearchActive] = useState(false)
  const mouseOver = useRef(false)
  const inputRef = useRef<HTMLInputElement | null>(null)

  // Close search input if mouse not over
  useEffect(() => {
    window.addEventListener("click", () => {
      if (mouseOver.current != true) {
        setSearchActive(false)
      }
    })
    return () => {
      window.removeEventListener("click", () => {
        if (!mouseOver.current != true) {
          setSearchActive(false)
        }
      })
    }
  }, [])

  useEffect(() => {
    if (searchActive && inputRef.current != null) {
      inputRef.current.focus()
    } 
  }, [searchActive])

  return (
    <div className="navbarSearchContainer" onMouseOver={() => mouseOver.current = true} onMouseOut={() => mouseOver.current = false}>
      <NavbarSearchInput active={searchActive} ref={inputRef}/>
      <Search color='black' onClick={() => {setSearchActive(!searchActive)}}/>
    </div>
  )
}
