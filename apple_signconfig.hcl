source = ["./build/darwinuniversal/txtwordle"]
bundle_id = "com.idolstarastronomer.txtwordle"

apple_id {
  password = "@env:AC_PASSWORD"
}

sign {
  application_identity = "Developer ID Application: Christopher De Vries (HV3HRV5DGR)"
}

zip {
  output_path = "dist/txtwordle-mac.zip"
}
