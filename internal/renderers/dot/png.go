package dot

import "os/exec"

//ToPng runs the dot command to convert a dot file into an image file
func ToPng(from, to, format string) error {
	_, err := exec.Command("sh", "-c", "dot -T"+format+" "+from+" -o "+to).Output()
	return err
}
