package xmlrpc

import (
	"encoding/xml"
	"testing"
)

func TestMarshal(t *testing.T) {
	expected_xml := `<methodResponse>
 <params>
  <param>
   <value>
    <array>
     <data>
      <value>
       <int>1</int>
      </value>
      <value>
       <string>Subscribed to [/rosout]</string>
      </value>
      <value>
       <array>
        <data>
         <value>
          <string>http://laptop-of-love:52764/</string>
         </value>
         <value>
          <string>http://laptop-of-love:41401/</string>
         </value>
        </data>
       </array>
      </value>
     </data>
    </array>
   </value>
  </param>
 </params>
</methodResponse>`
	one := 1
	str0 := "Subscribed to [/rosout]"
	str1 := "http://laptop-of-love:52764/"
	str2 := "http://laptop-of-love:41401/"
	data := MethodResponse{Params: []Param{Param{Val: Value{Array: &[]Value{
		Value{Int: &one},
		Value{String: &str0},
		Value{Array: &[]Value{
			Value{String: &str1},
			Value{String: &str2}}}}}}}}

	marshalled_xml, err := xml.MarshalIndent(data, "", " ")

	if err != nil {
		t.Errorf("Err != nil: %v\n", err)
	}
	if expected_xml != string(marshalled_xml) {
		t.Errorf("expected_xml != marshalled_xml:\nexpected=\n%v\n\nmarshalled=\n%v\n\n",
			expected_xml, string(marshalled_xml))
	}
}

func TestUnmarshal(t *testing.T) {
	input_xml := `<methodResponse>
 <params>
  <param>
   <value>
    <array>
     <data>
      <value>
       <int>1</int>
      </value>
      <value>
       <string>Subscribed to [/rosout]</string>
      </value>
      <value>
       <array>
        <data>
         <value>
          <string>http://laptop-of-love:52764/</string>
         </value>
         <value>
          <string>http://laptop-of-love:41401/</string>
         </value>
        </data>
       </array>
      </value>
     </data>
    </array>
   </value>
  </param>
 </params>
</methodResponse>`
	var unmarshalled_data MethodResponse
	err := xml.Unmarshal([]byte(input_xml), &unmarshalled_data)

	if err != nil {
		t.Errorf("Err != nil: %v\n", err)
	}
	array := *unmarshalled_data.Params[0].Val.Array
	if *array[0].Int != 1 {
		t.Errorf("Params[0].Val.Array[0].Int* != 1")
	}
	if *array[1].String != "Subscribed to [/rosout]" {
		t.Errorf("Params[0].Val.Array[1].String* != 'Subscribed to [/rosout]'")
	}
	inner_array := *array[2].Array
	if *inner_array[0].String != "http://laptop-of-love:52764/" {
		t.Errorf("*inner_array[0].String != 'http://laptop-of-love:52764/'")
	}
	if *inner_array[1].String != "http://laptop-of-love:41401/" {
		t.Errorf("*inner_array[0].String != 'http://laptop-of-love:41401/'")
	}
	if len(array) != 3 {
		t.Errorf("len(array) != 3, len(array) = %d\n", len(array))
	}
	if len(inner_array) != 2 {
		t.Errorf("len(inner_array) != 2, len(inner_array) = %d\n", len(inner_array))
	}
}
