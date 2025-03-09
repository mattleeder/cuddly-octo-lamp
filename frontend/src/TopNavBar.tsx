import React from 'react';
import { Dropdown, DropdownItem } from "./Dropdown.tsx"
import { ToggleDropdown, ToggleDropdownItem } from './ToggleDropdown.tsx';
import { Settings } from 'lucide-react';
import { NavbarSearch } from './NavSearch.tsx';

export function TopNavBar() {
  return (
    <nav className='topNavBar'>
      <div className='navBarContainer left'>
        <a href='#'>BurrChess</a>

        <Dropdown title='Play' titleHref='#'>
          <DropdownItem href="#">Test</DropdownItem>
        </Dropdown>

        <Dropdown title='Watch' titleHref='#'>
          <DropdownItem href="#">All</DropdownItem>
          <DropdownItem href="#">Bullet</DropdownItem>
          <DropdownItem href="#">Blitz</DropdownItem>
          <DropdownItem href="#">Rapid</DropdownItem>
          <DropdownItem href="#">Classical</DropdownItem>
        </Dropdown>
      </div>

      <div className='navBarContainer right'>
        <NavbarSearch />

        <a href='#'>Sign In</a>

        <ToggleDropdown title={<Settings />}>
          <ToggleDropdownItem href='#'>Settings</ToggleDropdownItem>
        </ToggleDropdown>
      </div>
    </nav>
  )
}