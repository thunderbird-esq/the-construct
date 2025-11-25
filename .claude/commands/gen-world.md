# Generate World Sections

Generate new world sections with rooms, NPCs, and items using procedural generation.

## Steps

1. Ask the user for generation parameters:
   - **Area type**: city, dungeon, wasteland, underground, Matrix construct, etc.
   - **Number of rooms**: (default: 10, range: 1-100)
   - **Difficulty level**: 1-10 (affects NPC strength and loot quality)

2. Based on the area type, generate appropriate:
   - Room descriptions (atmospheric, thematic)
   - Room connections (exits)
   - NPCs with appropriate stats for difficulty level
   - Items and loot tables
   - Quest opportunities

3. Create JSON structure compatible with `data/world.json` format:
   ```json
   {
     "room_id": {
       "ID": "room_id",
       "Description": "...",
       "Exits": {...},
       "Symbol": ".",
       "Color": "white",
       "Items": [...],
       "NPCs": [...]
     }
   }
   ```

4. Output the generated JSON

5. Ask if the user wants to:
   - Save to data/world.json (manual merge)
   - Generate more rooms
   - Adjust parameters

## Example Generation

For a "cyberpunk city, 5 rooms, difficulty 3":
- Dark alleyways with neon lighting
- Street vendors (friendly NPCs)
- Riot cops (hostile, difficulty-appropriate stats)
- Cyberdeck fragments, credsticks (loot)
- Connected via north/south/east/west exits

## Notes

- Generated content follows Matrix theme
- NPCs have appropriate difficulty scaling
- Loot quality matches difficulty level
- Room descriptions are atmospheric and varied
