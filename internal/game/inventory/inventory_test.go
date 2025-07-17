package inventory

import (
	"testing"

	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

func init() {
	// テスト用のログ初期化
	logger.Setup()
}

func TestNewInventory(t *testing.T) {
	inv := NewInventory()

	if inv == nil {
		t.Fatal("NewInventory() returned nil")
	}

	if len(inv.Items) != 0 {
		t.Errorf("New inventory should be empty, got %d items", len(inv.Items))
	}

	if inv.IsEmpty() != true {
		t.Error("New inventory should be empty")
	}

	if inv.IsFull() != false {
		t.Error("New inventory should not be full")
	}
}

func TestInventoryAddItem(t *testing.T) {
	inv := NewInventory()
	testItem := item.NewItem(0, 0, item.ItemWeapon, "Test Sword", 100)

	// アイテム追加
	result := inv.AddItem(testItem)
	if !result {
		t.Error("AddItem() should return true for successful addition")
	}

	if len(inv.Items) != 1 {
		t.Errorf("Inventory should have 1 item, got %d", len(inv.Items))
	}

	if inv.Items[0] != testItem {
		t.Error("Added item does not match expected item")
	}

	if inv.IsEmpty() {
		t.Error("Inventory should not be empty after adding item")
	}
}

func TestInventoryAddItemToFull(t *testing.T) {
	inv := NewInventory()

	// インベントリを満杯にする
	for i := 0; i < MaxInventorySize; i++ {
		testItem := item.NewItem(0, 0, item.ItemWeapon, "Test Item", 50)
		if !inv.AddItem(testItem) {
			t.Fatalf("Failed to add item %d to inventory", i)
		}
	}

	if !inv.IsFull() {
		t.Error("Inventory should be full")
	}

	// 満杯のインベントリにアイテムを追加しようとする
	extraItem := item.NewItem(0, 0, item.ItemArmor, "Extra Item", 75)
	result := inv.AddItem(extraItem)

	if result {
		t.Error("AddItem() should return false when inventory is full")
	}

	if len(inv.Items) != MaxInventorySize {
		t.Errorf("Inventory should have %d items, got %d", MaxInventorySize, len(inv.Items))
	}
}

func TestInventoryRemoveItem(t *testing.T) {
	inv := NewInventory()
	item1 := item.NewItem(0, 0, item.ItemWeapon, "Sword", 100)
	item2 := item.NewItem(0, 0, item.ItemArmor, "Armor", 150)
	item3 := item.NewItem(0, 0, item.ItemPotion, "Potion", 25)

	inv.AddItem(item1)
	inv.AddItem(item2)
	inv.AddItem(item3)

	// 中間のアイテムを削除
	inv.RemoveItem(1)

	if len(inv.Items) != 2 {
		t.Errorf("Inventory should have 2 items after removal, got %d", len(inv.Items))
	}

	if inv.Items[0] != item1 {
		t.Error("First item should remain unchanged")
	}

	if inv.Items[1] != item3 {
		t.Error("Third item should move to second position")
	}
}

func TestInventoryGetItem(t *testing.T) {
	inv := NewInventory()
	testItem := item.NewItem(0, 0, item.ItemWeapon, "Test Sword", 100)
	inv.AddItem(testItem)

	// 有効なインデックス
	retrieved := inv.GetItem(0)
	if retrieved != testItem {
		t.Error("GetItem() should return the correct item")
	}

	// 無効なインデックス
	invalidItem := inv.GetItem(10)
	if invalidItem != nil {
		t.Error("GetItem() should return nil for invalid index")
	}

	// 負のインデックス
	negativeItem := inv.GetItem(-1)
	if negativeItem != nil {
		t.Error("GetItem() should return nil for negative index")
	}
}

func TestInventoryGetInventoryListing(t *testing.T) {
	inv := NewInventory()
	
	// 空のインベントリ
	listing := inv.GetInventoryListing(nil)
	expectedEmpty := []string{"Your pack is empty."}
	
	if len(listing) != len(expectedEmpty) {
		t.Errorf("Empty inventory listing length = %d, want %d", len(listing), len(expectedEmpty))
	}
	
	if listing[0] != expectedEmpty[0] {
		t.Errorf("Empty inventory message = %q, want %q", listing[0], expectedEmpty[0])
	}

	// アイテムを追加
	item1 := item.NewItem(0, 0, item.ItemWeapon, "Iron Sword", 120)
	item2 := item.NewItem(0, 0, item.ItemPotion, "Health Potion", 30)
	inv.AddItem(item1)
	inv.AddItem(item2)

	listing = inv.GetInventoryListing(nil)
	
	if len(listing) != 3 {
		t.Errorf("Inventory listing length = %d, want 3", len(listing))
	}

	// ヘッダーとアイテムの順序と形式をチェック
	expectedHeader := "Current inventory:"
	expectedFormat1 := "a) Iron Sword"
	expectedFormat2 := "b) Health Potion"

	if listing[0] != expectedHeader {
		t.Errorf("Header listing = %q, want %q", listing[0], expectedHeader)
	}

	if listing[1] != expectedFormat1 {
		t.Errorf("First item listing = %q, want %q", listing[1], expectedFormat1)
	}

	if listing[2] != expectedFormat2 {
		t.Errorf("Second item listing = %q, want %q", listing[2], expectedFormat2)
	}
}

func TestNewEquipment(t *testing.T) {
	eq := NewEquipment()

	if eq == nil {
		t.Fatal("NewEquipment() returned nil")
	}

	if eq.Weapon != nil {
		t.Error("New equipment should have no weapon")
	}

	if eq.Armor != nil {
		t.Error("New equipment should have no armor")
	}

	if eq.RingLeft != nil {
		t.Error("New equipment should have no left ring")
	}

	if eq.RingRight != nil {
		t.Error("New equipment should have no right ring")
	}
}

func TestEquipmentEquipItem(t *testing.T) {
	
	tests := []struct {
		name      string
		item      *item.Item
		wantEquip bool
		checkSlot func(*Equipment) *item.Item
	}{
		{
			name:      "武器装備",
			item:      item.NewItem(0, 0, item.ItemWeapon, "Sword", 100),
			wantEquip: true,
			checkSlot: func(e *Equipment) *item.Item { return e.Weapon },
		},
		{
			name:      "防具装備",
			item:      item.NewItem(0, 0, item.ItemArmor, "Armor", 150),
			wantEquip: true,
			checkSlot: func(e *Equipment) *item.Item { return e.Armor },
		},
		{
			name:      "指輪装備",
			item:      item.NewItem(0, 0, item.ItemRing, "Ring", 75),
			wantEquip: true,
			checkSlot: func(e *Equipment) *item.Item { return e.RingLeft },
		},
		{
			name:      "装備不可アイテム",
			item:      item.NewItem(0, 0, item.ItemPotion, "Potion", 25),
			wantEquip: false,
			checkSlot: func(e *Equipment) *item.Item { return nil },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq := NewEquipment() // 新しい装備を作成
			result := eq.EquipItem(tt.item)

			if result != tt.wantEquip {
				t.Errorf("EquipItem() = %v, want %v", result, tt.wantEquip)
			}

			if tt.wantEquip {
				equipped := tt.checkSlot(eq)
				if equipped != tt.item {
					t.Errorf("Item was not equipped in correct slot")
				}
			}
		})
	}
}

func TestEquipmentUnequipItem(t *testing.T) {
	eq := NewEquipment()
	weapon := item.NewItem(0, 0, item.ItemWeapon, "Sword", 100)
	
	// 武器を装備
	eq.EquipItem(weapon)
	
	// 武器を外す
	unequipped := eq.UnequipItem("weapon")
	if unequipped != weapon {
		t.Error("UnequipItem() should return the unequipped weapon")
	}
	
	if eq.Weapon != nil {
		t.Error("Weapon slot should be empty after unequipping")
	}
	
	// 空のスロットから外そうとする
	nothing := eq.UnequipItem("weapon")
	if nothing != nil {
		t.Error("UnequipItem() should return nil for empty slot")
	}
	
	// 無効なスロット名
	invalid := eq.UnequipItem("invalid")
	if invalid != nil {
		t.Error("UnequipItem() should return nil for invalid slot")
	}
}

func TestEquipmentGetEquippedNames(t *testing.T) {
	eq := NewEquipment()
	
	// 空の装備
	weapon, armor, ringLeft, ringRight := eq.GetEquippedNames()
	expected := noneEquipped
	
	if weapon != expected || armor != expected || ringLeft != expected || ringRight != expected {
		t.Errorf("Empty equipment names = (%q, %q, %q, %q), want all %q", 
			weapon, armor, ringLeft, ringRight, expected)
	}
	
	// アイテムを装備
	weaponItem := item.NewItem(0, 0, item.ItemWeapon, "Iron Sword", 120)
	armorItem := item.NewItem(0, 0, item.ItemArmor, "Chain Mail", 200)
	eq.EquipItem(weaponItem)
	eq.EquipItem(armorItem)
	
	weapon, armor, ringLeft, ringRight = eq.GetEquippedNames()
	
	if weapon != "Iron Sword" {
		t.Errorf("Weapon name = %q, want %q", weapon, "Iron Sword")
	}
	
	if armor != "Chain Mail" {
		t.Errorf("Armor name = %q, want %q", armor, "Chain Mail")
	}
	
	if ringLeft != expected {
		t.Errorf("Ring left name = %q, want %q", ringLeft, expected)
	}
	
	if ringRight != expected {
		t.Errorf("Ring right name = %q, want %q", ringRight, expected)
	}
}

func TestEquipmentGetAttackBonus(t *testing.T) {
	eq := NewEquipment()
	
	// 初期状態
	if bonus := eq.GetAttackBonus(); bonus != 0 {
		t.Errorf("Initial attack bonus = %d, want 0", bonus)
	}
	
	// 武器を装備（Value=100なので攻撃ボーナス=10）
	weapon := item.NewItem(0, 0, item.ItemWeapon, "Sword", 100)
	eq.EquipItem(weapon)
	
	// 武器ボーナス = Value / 10 = 100 / 10 = 10
	expectedBonus := 10
	if bonus := eq.GetAttackBonus(); bonus != expectedBonus {
		t.Errorf("Attack bonus with weapon = %d, want %d", bonus, expectedBonus)
	}
}

func TestEquipmentGetDefenseBonus(t *testing.T) {
	eq := NewEquipment()
	
	// 初期状態
	if bonus := eq.GetDefenseBonus(); bonus != 0 {
		t.Errorf("Initial defense bonus = %d, want 0", bonus)
	}
	
	// 防具を装備（Value=200なので防御ボーナス=20）
	armor := item.NewItem(0, 0, item.ItemArmor, "Chain Mail", 200)
	eq.EquipItem(armor)
	
	// 防具ボーナス = Value / 10 = 200 / 10 = 20
	expectedBonus := 20
	if bonus := eq.GetDefenseBonus(); bonus != expectedBonus {
		t.Errorf("Defense bonus with armor = %d, want %d", bonus, expectedBonus)
	}
}